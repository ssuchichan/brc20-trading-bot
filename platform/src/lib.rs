extern crate core;
pub mod robot;

extern crate dotenv;
use anyhow::{anyhow, Result};
use base64::engine;
use base64::{engine::general_purpose::URL_SAFE, Engine as _};
use core::slice;
use dotenv::dotenv;
use finutils::txn_builder::TransactionBuilder;
use finutils::txn_builder::TransferOperationBuilder;
use globutils::wallet;
use globutils::wallet::restore_keypair_from_seckey_base64;
use ledger::data_model::BLACK_HOLE_PUBKEY;
use ledger::data_model::TX_FEE_MIN_V0;
use ledger::data_model::{b64dec, TransferType, TxoRef, TxoSID, Utxo, ASSET_TYPE_FRA};
use rand::Rng;
use robot::Robot;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use sha2::{Digest, Sha256};
use sqlx::postgres::PgPoolOptions;
use std::collections::HashMap;
use std::ffi::CString;
use std::os::raw::c_char;
use zei::serialization::ZeiFromToBytes;
use zei::xfr::asset_record::{open_blind_asset_record, AssetRecordType};
use zei::xfr::sig::XfrPublicKey;
use zei::xfr::structs::{AssetRecordTemplate, OwnerMemo};

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Memo {
    p: String,
    op: String,
    tick: String,
    amt: String,
}

impl Memo {
    fn new(p: String, op: String, tick: String, amt: String) -> Self {
        Self { p, op, tick, amt }
    }
}

#[no_mangle]
pub extern "C" fn get_tx_str(
    from_sig_ptr: *mut u8, // 发送者私钥
    from_sig_len: u32,
    fra_receiver_ptr: *mut u8, // 接收者公钥
    fra_receiver_len: u32,
    to_ptr: *mut u8, // 接收者公钥
    to_len: u32,
    trans_amount_ptr: *mut u8, // 数量
    trans_amount_len: u32,
    url_ptr: *mut u8, // 节点RPC
    url_len: u32,
    tick_ptr: *mut u8, // token
    tick_len: u8,
    fra_price_ptr: *mut u8, // price
    fra_price_len: u32,
    brc_type_ptr: *mut u8, // type
    brc_type_len: u32,
) -> *const c_char {
    let from_key = unsafe { slice::from_raw_parts(from_sig_ptr, from_sig_len as usize) };
    let from_key_str = std::str::from_utf8(from_key).unwrap();
    let from_key_pair = restore_keypair_from_seckey_base64(from_key_str).unwrap();

    let to_pub_key = unsafe { slice::from_raw_parts(to_ptr, to_len as usize) };
    let to_dec = b64dec(to_pub_key).unwrap();
    let to = XfrPublicKey::zei_from_bytes(to_dec.as_slice()).unwrap();

    let fra_receiver_key =
        unsafe { slice::from_raw_parts(fra_receiver_ptr, fra_receiver_len as usize) };
    let fra_dec = b64dec(fra_receiver_key).unwrap();
    let fra_receiver = XfrPublicKey::zei_from_bytes(fra_dec.as_slice()).unwrap();

    let tick = unsafe { slice::from_raw_parts(tick_ptr, tick_len as usize) };

    let trans_amount =
        unsafe { slice::from_raw_parts(trans_amount_ptr, trans_amount_len as usize) };
    let trans_amount_str = std::str::from_utf8(trans_amount).unwrap();
    let amount = trans_amount_str.parse::<u64>().unwrap();

    let url = unsafe { slice::from_raw_parts(url_ptr, url_len as usize) };
    let url_str = std::str::from_utf8(url).unwrap();

    let brc_type = unsafe { slice::from_raw_parts(brc_type_ptr, brc_type_len as usize) };
    let brc_type_str = std::str::from_utf8(brc_type).unwrap();

    let fra_price = unsafe { slice::from_raw_parts(fra_price_ptr, fra_price_len as usize) };
    let fra_amount_str = std::str::from_utf8(fra_price).unwrap();
    let price = fra_amount_str.parse::<u64>().unwrap();

    let payment = amount * price;

    let asset_record_type = AssetRecordType::from_flags(false, false);

    let mut op = TransferOperationBuilder::new();

    // build input
    let mut input_amount = 0;
    let mut t_amout;
    let utxos = get_owned_utxos_x(
        url_str,
        wallet::public_key_to_base64(from_key_pair.get_pk_ref()).as_str(),
    )
    .unwrap();
    for (sid, (utxo, owner_memo)) in utxos.into_iter() {
        let oar = open_blind_asset_record(&utxo.0.record, &owner_memo, &from_key_pair).unwrap();
        if oar.asset_type != ASSET_TYPE_FRA {
            continue;
        }
        t_amout = oar.amount;
        input_amount += t_amout;

        if t_amout != 0 {
            op.add_input(TxoRef::Absolute(sid), oar, None, None, t_amout)
                .unwrap();
            if input_amount > payment + TX_FEE_MIN_V0 {
                // if input bigger than trans amount
                break;
            }
        }
    }

    if input_amount < payment + TX_FEE_MIN_V0 {
        return CString::new("").unwrap().into_raw();
    }

    let memo_struct = Memo::new(
        "brc-20".to_string(),
        brc_type_str.to_string(),
        std::str::from_utf8(tick).unwrap().to_string(),
        trans_amount_str.to_string(),
    );
    let memo = serde_json::to_string(&memo_struct).unwrap();
    let template =
        AssetRecordTemplate::with_no_asset_tracing(0, ASSET_TYPE_FRA, asset_record_type, to);

    let template_from = AssetRecordTemplate::with_no_asset_tracing(
        input_amount - TX_FEE_MIN_V0 - payment,
        ASSET_TYPE_FRA,
        asset_record_type,
        from_key_pair.get_pk(),
    );

    let template_fee = AssetRecordTemplate::with_no_asset_tracing(
        TX_FEE_MIN_V0,
        ASSET_TYPE_FRA,
        asset_record_type,
        *BLACK_HOLE_PUBKEY,
    );

    let receive_fra = AssetRecordTemplate::with_no_asset_tracing(
        payment,
        ASSET_TYPE_FRA,
        asset_record_type,
        fra_receiver,
    );
    // build output
    let trans_build = op
        .add_output(&template_fee, None, None, None, None)
        .and_then(|b| b.add_output(&template, None, None, None, Some(memo)))
        .and_then(|b| b.add_output(&template_from, None, None, None, None))
        .and_then(|b| b.add_output(&receive_fra, None, None, None, None))
        .and_then(|b| b.create(TransferType::Standard))
        .and_then(|b| b.sign(&from_key_pair))
        .and_then(|b| b.transaction())
        .unwrap();

    let mut builder: TransactionBuilder = get_transaction_builder(url_str).unwrap();

    let tx = builder
        .add_operation(trans_build)
        .sign_to_map(&from_key_pair)
        .clone()
        .take_transaction();

    let tx_str = serde_json::to_string(&tx).unwrap();
    let c_string = CString::new(tx_str).unwrap();
    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn get_transfer_tx_str(
    from_sig_ptr: *mut u8,
    from_sig_len: u32,
    fra_receiver_ptr: *mut u8,
    fra_receiver_len: u32,
    fra_price_ptr: *mut u8,
    fra_price_len: u32,
    url_ptr: *mut u8,
    url_len: u32,
) -> *const c_char {
    let from_key = unsafe { slice::from_raw_parts(from_sig_ptr, from_sig_len as usize) };
    let fra_receiver_key =
        unsafe { slice::from_raw_parts(fra_receiver_ptr, fra_receiver_len as usize) };

    let url = unsafe { slice::from_raw_parts(url_ptr, url_len as usize) };
    let url_str = std::str::from_utf8(url).unwrap();

    let from_key_str = std::str::from_utf8(from_key).unwrap();
    //let from = wallet::restore_keypair_from_mnemonic_default(from_key_str).unwrap();
    let from = restore_keypair_from_seckey_base64(from_key_str).unwrap();

    let fra_dec = b64dec(fra_receiver_key).unwrap();
    let fra_receiver = XfrPublicKey::zei_from_bytes(fra_dec.as_slice()).unwrap();

    let fra_amount = unsafe { slice::from_raw_parts(fra_price_ptr, fra_price_len as usize) };
    let fra_amount_str = std::str::from_utf8(fra_amount).unwrap();
    let num = fra_amount_str.parse::<f64>().unwrap();
    let fra_price = (num * 1000000.0) as u64;

    let asset_record_type = AssetRecordType::from_flags(false, false);

    let mut op = TransferOperationBuilder::new();

    // build input
    let mut input_amount = 0;
    let mut t_amout;
    let utxos = get_owned_utxos_x(
        url_str,
        wallet::public_key_to_base64(from.get_pk_ref()).as_str(),
    )
    .unwrap();
    for (sid, (utxo, owner_memo)) in utxos.into_iter() {
        let oar = open_blind_asset_record(&utxo.0.record, &owner_memo, &from).unwrap();
        if oar.asset_type != ASSET_TYPE_FRA {
            continue;
        }
        t_amout = oar.amount;
        input_amount += t_amout;

        if t_amout != 0 {
            op.add_input(TxoRef::Absolute(sid), oar, None, None, t_amout)
                .unwrap();
            if input_amount > fra_price + TX_FEE_MIN_V0 {
                // if input big than trans amount
                break;
            }
        }
    }

    if input_amount < fra_price + TX_FEE_MIN_V0 {
        return CString::new("").unwrap().into_raw();
    }

    // 找零
    let template_from = AssetRecordTemplate::with_no_asset_tracing(
        input_amount - TX_FEE_MIN_V0 - fra_price,
        ASSET_TYPE_FRA,
        asset_record_type,
        from.get_pk(),
    );

    // 手续费
    let template_fee = AssetRecordTemplate::with_no_asset_tracing(
        TX_FEE_MIN_V0,
        ASSET_TYPE_FRA,
        asset_record_type,
        *BLACK_HOLE_PUBKEY,
    );

    // 转账
    let receive_fra = AssetRecordTemplate::with_no_asset_tracing(
        fra_price,
        ASSET_TYPE_FRA,
        asset_record_type,
        fra_receiver,
    );

    let trans_build = op
        .add_output(&template_fee, None, None, None, None)
        .and_then(|b| b.add_output(&template_from, None, None, None, None))
        .and_then(|b| b.add_output(&receive_fra, None, None, None, None))
        .and_then(|b| b.create(TransferType::Standard))
        .and_then(|b| b.sign(&from))
        .and_then(|b| b.transaction())
        .unwrap();

    let mut builder: TransactionBuilder = get_transaction_builder(url_str).unwrap();

    let tx = builder
        .add_operation(trans_build)
        .sign_to_map(&from)
        .clone()
        .take_transaction();

    let tx_str = serde_json::to_string(&tx).unwrap();
    let c_string = CString::new(tx_str).unwrap();
    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn get_seq_id(url_ptr: *mut u8, url_len: u32) -> u64 {
    let url = unsafe { slice::from_raw_parts(url_ptr, url_len as usize) };
    let url_str = std::str::from_utf8(url).unwrap();
    let result = get_transaction_builder(url_str).unwrap();
    result.get_seq_id()
}

fn get_transaction_builder(url: &str) -> Result<TransactionBuilder> {
    let url = format!("{}/global_state", url);
    attohttpc::get(&url)
        .send()
        .and_then(|resp| resp.error_for_status())
        .and_then(|resp| resp.bytes())
        .map_err(|e| anyhow!("{:?}", e))
        .and_then(|bytes| {
            serde_json::from_slice::<(Value, u64, Value)>(&bytes).map_err(|e| anyhow!("{:?}", e))
        })
        .map(|resp| TransactionBuilder::from_seq_id(resp.1))
}

fn get_owned_utxos_x(
    url: &str,
    pubkey: &str,
) -> Result<HashMap<TxoSID, (Utxo, Option<OwnerMemo>)>> {
    let url = format!("{}/owned_utxos/{}", url, pubkey);

    attohttpc::get(url)
        .send()
        .and_then(|resp| resp.bytes())
        .map_err(|e| anyhow! {"{:?}", e})
        .and_then(|b| {
            serde_json::from_slice::<HashMap<TxoSID, (Utxo, Option<OwnerMemo>)>>(&b)
                .map_err(|e| anyhow!("{:?}", e))
        })
}

#[no_mangle]
pub extern "C" fn send_tx(
    tx_prt: *mut u8,
    tx_len: u32,
    url_ptr: *mut u8,
    url_len: u32,
) -> *const c_char {
    let tx_u8 = unsafe { slice::from_raw_parts(tx_prt, tx_len as usize) };
    let tx = std::str::from_utf8(tx_u8).unwrap();
    let tx_bytes = hex::decode(tx.strip_prefix("0x").unwrap_or(tx)).unwrap();
    let url = unsafe { slice::from_raw_parts(url_ptr, url_len as usize) };
    let url_str = std::str::from_utf8(url).unwrap();
    let txn_b64 = URL_SAFE.encode(&tx_bytes);

    let json_rpc =
        format!(
                "{{\"jsonrpc\":\"2.0\",\"id\":\"anything\",\"method\":\"broadcast_tx_sync\",\"params\": {{\"tx\": \"{}\"}}}}",
                &txn_b64
            );

    attohttpc::post(url_str)
        .header(attohttpc::header::CONTENT_TYPE, "application/json")
        .text(json_rpc)
        .send()
        .and_then(|v| v.error_for_status())
        .map(|v| println!("{:?}", v))
        .unwrap();

    let tx_hash = Sha256::digest(&tx_bytes);
    let result = hex::encode(tx_hash);
    let c_string = CString::new(result).unwrap();
    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn generate_mnemonic_default() -> *const c_char {
    let r = wallet::generate_mnemonic_default();
    let c_string = CString::new(r).unwrap();
    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn mnemonic_to_bench32(from_sig_ptr: *mut u8, from_sig_len: u32) -> *const c_char {
    let from_key = unsafe { slice::from_raw_parts(from_sig_ptr, from_sig_len as usize) };
    let from_key_str = std::str::from_utf8(from_key).unwrap();
    let from = wallet::restore_keypair_from_mnemonic_default(from_key_str).unwrap();
    let pub_key = from.get_pk();
    let from_bench32 = wallet::public_key_to_bech32(&pub_key);
    let c_string = CString::new(from_bench32).unwrap();

    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn mnemonic_to_private_key(
    mnemonic_ptr: *mut u8,
    mnemonic_len: u32,
) -> *const c_char {
    let mnemonic = unsafe { slice::from_raw_parts(mnemonic_ptr, mnemonic_len as usize) };
    let mnemonic_str = std::str::from_utf8(mnemonic).unwrap();
    let key_pair = wallet::restore_keypair_from_mnemonic_default(mnemonic_str).unwrap();
    let sk = key_pair.get_sk().zei_to_bytes();
    let sk_base64 = engine::general_purpose::URL_SAFE.encode(sk);
    let c_string = CString::new(sk_base64).unwrap();

    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn mnemonic_to_public_key(
    mnemonic_ptr: *mut u8,
    mnemonic_len: u32,
) -> *const c_char {
    let mnemonic = unsafe { slice::from_raw_parts(mnemonic_ptr, mnemonic_len as usize) };
    let mnemonic_str = std::str::from_utf8(mnemonic).unwrap();
    let key_pair = wallet::restore_keypair_from_mnemonic_default(mnemonic_str).unwrap();
    let sk = key_pair.get_pk().zei_to_bytes();
    let sk_base64 = engine::general_purpose::URL_SAFE.encode(sk);
    let c_string = CString::new(sk_base64).unwrap();

    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn generate_private_key() -> *const c_char {
    let mnemonic = wallet::generate_mnemonic_default();
    let key_pair = wallet::restore_keypair_from_mnemonic_default(&mnemonic).unwrap();
    let sk = key_pair.get_sk().zei_to_bytes();
    let sk_base64 = engine::general_purpose::URL_SAFE.encode(sk);
    let c_string = CString::new(sk_base64).unwrap();

    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn private_key_to_bech32(sk_base64_ptr: *mut u8, key_len: u32) -> *const c_char {
    let sk_bytes = unsafe { slice::from_raw_parts(sk_base64_ptr, key_len as usize) };
    let sk_str = std::str::from_utf8(sk_bytes).unwrap();
    let key_pair = wallet::restore_keypair_from_seckey_base64(sk_str).unwrap();
    let pubkey_bech32 = wallet::public_key_to_bech32(&key_pair.get_pk());
    let c_string = CString::new(pubkey_bech32).unwrap();

    c_string.into_raw()
}

#[derive(Serialize)]
struct RobotInitAmount {
    pub account: String,
    pub amount: u64,
}

impl RobotInitAmount {
    fn new(account: String, amount: u64) -> Self {
        Self { account, amount }
    }
}

#[no_mangle]
pub extern "C" fn get_send_robot_batch_tx(
    from_sig_ptr: *mut u8,
    from_sig_len: u32,
    url_ptr: *mut u8,
    url_len: u32,
) -> *const c_char {
    dotenv().ok();
    let from_key = unsafe { slice::from_raw_parts(from_sig_ptr, from_sig_len as usize) };
    let url = unsafe { slice::from_raw_parts(url_ptr, url_len as usize) };
    let url_str = std::str::from_utf8(url).unwrap();

    let from_key_str = std::str::from_utf8(from_key).unwrap();
    let from = wallet::restore_keypair_from_mnemonic_default(from_key_str).unwrap();

    let accounts_result = tokio::runtime::Runtime::new().unwrap().block_on(async {
        let db_user = std::env::var("DBUSER").unwrap();
        let db_password = std::env::var("PASSWORD").unwrap();
        let db_host = std::env::var("HOST").unwrap();
        let db_port = std::env::var("PORT").unwrap();
        let db_name = std::env::var("DBName").unwrap();

        let uri = format!(
            "postgres://{}:{}@{}:{}/{}",
            db_user, db_password, db_host, db_port, db_name
        );

        let pg_pool = PgPoolOptions::new().connect(&uri).await.unwrap();
        Robot::all_accounts(&pg_pool)
            .await
            .unwrap_or_else(|_| vec![])
    });

    //     if accounts_result.len() != 200 {
    //         return CString::new("").unwrap().into_raw();
    //     }

    let mut robot_users: Vec<RobotInitAmount> = Vec::with_capacity(accounts_result.len() as usize);
    let mut fra_price_total = 0;
    let mut rng = rand::thread_rng();
    for account in accounts_result {
        let rand_num = rng.gen_range(1..11);
        fra_price_total += rand_num;
        robot_users.push(RobotInitAmount::new(account, rand_num as u64));
    }

    let fra_price = fra_price_total * 1000000;

    let asset_record_type = AssetRecordType::from_flags(false, false);

    let mut op = TransferOperationBuilder::new();

    // build input
    let mut input_amount = 0;
    let mut t_amout;
    let utxos = get_owned_utxos_x(
        url_str,
        wallet::public_key_to_base64(from.get_pk_ref()).as_str(),
    )
    .unwrap();
    for (sid, (utxo, owner_memo)) in utxos.into_iter() {
        let oar = open_blind_asset_record(&utxo.0.record, &owner_memo, &from).unwrap();
        if oar.asset_type != ASSET_TYPE_FRA {
            continue;
        }
        t_amout = oar.amount;
        input_amount += t_amout;

        if t_amout != 0 {
            op.add_input(TxoRef::Absolute(sid), oar, None, None, t_amout)
                .unwrap();
            if input_amount > fra_price + TX_FEE_MIN_V0 {
                // if input big than trans amount
                break;
            }
        }
    }

    if input_amount < fra_price + TX_FEE_MIN_V0 {
        return CString::new("").unwrap().into_raw();
    }

    // 找零
    let template_from = AssetRecordTemplate::with_no_asset_tracing(
        input_amount - TX_FEE_MIN_V0 - fra_price,
        ASSET_TYPE_FRA,
        asset_record_type,
        from.get_pk(),
    );

    // 手续费
    let template_fee = AssetRecordTemplate::with_no_asset_tracing(
        TX_FEE_MIN_V0,
        ASSET_TYPE_FRA,
        asset_record_type,
        *BLACK_HOLE_PUBKEY,
    );

    op.add_output(&template_fee, None, None, None, None)
        .and_then(|b| b.add_output(&template_from, None, None, None, None))
        .unwrap();

    for out in robot_users {
        // 转账
        let receive_fra = AssetRecordTemplate::with_no_asset_tracing(
            out.amount * 1000000,
            ASSET_TYPE_FRA,
            asset_record_type,
            wallet::public_key_from_bech32(&out.account).unwrap(),
        );
        op.add_output(&receive_fra, None, None, None, None).unwrap();
    }

    let trans_build = op
        .create(TransferType::Standard)
        .and_then(|b| b.sign(&from))
        .and_then(|b| b.transaction())
        .unwrap();

    let mut builder: TransactionBuilder = get_transaction_builder(url_str).unwrap();

    let tx = builder
        .add_operation(trans_build)
        .sign_to_map(&from)
        .clone()
        .take_transaction();

    let tx_str = serde_json::to_string(&tx).unwrap();
    let c_string = CString::new(tx_str).unwrap();
    c_string.into_raw()
}

#[no_mangle]
pub extern "C" fn get_user_fra_balance(
    from_sig_ptr: *mut u8,
    from_sig_len: u32,
    url_ptr: *mut u8,
    url_len: u32,
) -> u64 {
    let from_key = unsafe { slice::from_raw_parts(from_sig_ptr, from_sig_len as usize) };
    let url = unsafe { slice::from_raw_parts(url_ptr, url_len as usize) };
    let url_str = std::str::from_utf8(url).unwrap();

    let from_key_str = std::str::from_utf8(from_key).unwrap();
    //let from = wallet::restore_keypair_from_mnemonic_default(from_key_str).unwrap();
    let from = restore_keypair_from_seckey_base64(from_key_str).unwrap();

    // build input
    let mut input_amount = 0;
    let utxos = get_owned_utxos_x(
        url_str,
        wallet::public_key_to_base64(from.get_pk_ref()).as_str(),
    )
    .unwrap();
    for (_, (utxo, owner_memo)) in utxos.into_iter() {
        let oar = open_blind_asset_record(&utxo.0.record, &owner_memo, &from).unwrap();
        if oar.asset_type != ASSET_TYPE_FRA {
            continue;
        }
        input_amount += oar.amount;
    }

    input_amount
}

#[cfg(test)]
mod tests {
    use crate::{
        generate_mnemonic_default, get_send_robot_batch_tx, get_transfer_tx_str, get_tx_str,
        get_user_fra_balance, robot::Robot, send_tx, Memo,
    };
    extern crate dotenv;
    use dotenv::dotenv;
    use globutils::wallet;
    use sqlx::postgres::PgPoolOptions;

    #[test]
    fn test_memo() {
        let memo_struct = Memo::new(
            "brc-20".to_string(),
            "transfer".to_string(),
            "ordi".to_string(),
            "1000".to_string(),
        );
        let memo = serde_json::to_string(&memo_struct).unwrap();
        println!("{}", memo);
        assert_eq!(
            "{\"p\":\"brc-20\",\"op\":\"transfer\",\"tick\":\"ordi\",\"amt\":\"1000\"}",
            memo
        )
    }

    #[test]
    fn test_env() {
        dotenv().ok();
        let mut from = std::env::var("CENTEREFROM").unwrap();

        let mut to = String::from("Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=");
        let mut receiver = String::from("Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=");
        let mut url = String::from("https://prod-testnet.prod.findora.org:8668");
        let mut trans = String::from("1111");
        let mut tick = String::from("only");
        let mut fra_amount = String::from("2.34");
        let mut brc_type = String::from("transfer");
        let a = get_tx_str(
            from.as_mut_ptr(),
            from.len() as u32,
            receiver.as_mut_ptr(),
            receiver.len() as u32,
            to.as_mut_ptr(),
            to.len() as u32,
            trans.as_mut_ptr(),
            trans.len() as u32,
            url.as_mut_ptr(),
            url.len() as u32,
            tick.as_mut_ptr(),
            tick.len() as u8,
            fra_amount.as_mut_ptr(),
            fra_amount.len() as u32,
            brc_type.as_mut_ptr(),
            brc_type.len() as u32,
        );
        let result = unsafe { std::ffi::CStr::from_ptr(a).to_str() };
        println!("result {:?}", result)
    }

    #[test]
    fn test_tranfer() {
        dotenv().ok();
        let mut from = std::env::var("CENTEREFROM").unwrap();
        let mut receiver = String::from("Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=");
        let mut url = String::from("https://prod-testnet.prod.findora.org:8668");
        let mut trans = String::from("2.34");
        let a = get_transfer_tx_str(
            from.as_mut_ptr(),
            from.len() as u32,
            receiver.as_mut_ptr(),
            receiver.len() as u32,
            trans.as_mut_ptr(),
            trans.len() as u32,
            url.as_mut_ptr(),
            url.len() as u32,
        );
        let result = unsafe { std::ffi::CStr::from_ptr(a).to_str() };
        println!("result {:?}", result)
    }

    #[test]
    fn test_send() {
        let mut tx = String::from("7b22626f6479223a7b226e6f5f7265706c61795f746f6b656e223a5b5b3131302c3230372c33332c3138362c3235342c3134372c3138362c3133375d2c31393132315d2c226f7065726174696f6e73223a5b7b225472616e736665724173736574223a7b22626f6479223a7b22696e70757473223a5b7b224162736f6c757465223a35373331347d5d2c22706f6c6963696573223a7b2276616c6964223a747275652c22696e707574735f74726163696e675f706f6c6963696573223a5b5b5d5d2c22696e707574735f7369675f636f6d6d69746d656e7473223a5b6e756c6c5d2c226f7574707574735f74726163696e675f706f6c6963696573223a5b5b5d2c5b5d2c5b5d5d2c226f7574707574735f7369675f636f6d6d69746d656e7473223a5b6e756c6c2c6e756c6c2c6e756c6c5d7d2c226f757470757473223a5b7b226964223a6e756c6c2c227265636f7264223a7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a223130303030227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a22414141414141414141414141414141414141414141414141414141414141414141414141414141414141413d227d7d2c7b226964223a6e756c6c2c227265636f7264223a7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a2231227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a226c30517567744e554b47443533574b6646364271364548457550775a63526e45766e5f5047343852336a673d227d2c226d656d6f223a227b5c22705c223a5c226272632d32305c222c5c226f705c223a5c227472616e736665725c222c5c227469636b5c223a5c226f7264695c222c5c22616d745c223a5c22313030305c227d227d2c7b226964223a6e756c6c2c227265636f7264223a7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a223437343839363233313039393935227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a22485a6e787750493550445f78705158314e714b54485871506448585658744765377951304a49334d5654733d227d7d5d2c227472616e73666572223a7b22696e70757473223a5b7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a223437343839363233313139393936227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a22485a6e787750493550445f78705158314e714b54485871506448585658744765377951304a49334d5654733d227d5d2c226f757470757473223a5b7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a223130303030227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a22414141414141414141414141414141414141414141414141414141414141414141414141414141414141413d227d2c7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a2231227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a226c30517567744e554b47443533574b6646364271364548457550775a63526e45766e5f5047343852336a673d227d2c7b22616d6f756e74223a7b224e6f6e436f6e666964656e7469616c223a223437343839363233313039393935227d2c2261737365745f74797065223a7b224e6f6e436f6e666964656e7469616c223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c305d7d2c227075626c69635f6b6579223a22485a6e787750493550445f78705158314e714b54485871506448585658744765377951304a49334d5654733d227d5d2c2270726f6f6673223a7b2261737365745f747970655f616e645f616d6f756e745f70726f6f66223a224e6f50726f6f66222c2261737365745f74726163696e675f70726f6f66223a7b2261737365745f747970655f616e645f616d6f756e745f70726f6f6673223a5b5d2c22696e707574735f6964656e746974795f70726f6f6673223a5b5b5d5d2c226f7574707574735f6964656e746974795f70726f6f6673223a5b5b5d2c5b5d2c5b5d5d7d7d2c2261737365745f74726163696e675f6d656d6f73223a5b5b5d2c5b5d2c5b5d2c5b5d5d2c226f776e6572735f6d656d6f73223a5b6e756c6c2c6e756c6c2c6e756c6c5d7d2c227472616e736665725f74797065223a225374616e64617264227d2c22626f64795f7369676e617475726573223a5b7b2261646472657373223a7b226b6579223a22485a6e787750493550445f78705158314e714b54485871506448585658744765377951304a49334d5654733d227d2c227369676e6174757265223a22594e2d473466437258334b5938614d464a78734e6a654b6135527266396c79584861453975773876583877455f557575706633736a6b67477530347573334e574f6d6f4167733042317a32624a31636f7356396344513d3d227d5d7d7d5d7d2c227075626b65795f7369676e5f6d6170223a7b22485a6e787750493550445f78705158314e714b54485871506448585658744765377951304a49334d5654733d223a2261556f7775334b7462634d446b4539592d4b6673646859617467596852613367475833774464505a595337394847634a6d617a6d567134337669384d6561584e4c446944333268324165716f4c506a31554c713742773d3d227d7d");
        let mut url = String::from("https://prod-testnet.prod.findora.org:26657");
        let a = send_tx(
            tx.as_mut_ptr(),
            tx.len() as u32,
            url.as_mut_ptr(),
            url.len() as u32,
        );
        let result = unsafe { std::ffi::CStr::from_ptr(a).to_str() };
        println!("{:?}", result)
    }

    #[test]
    fn test_generate_mnemonic() {
        let a = generate_mnemonic_default();
        let result = unsafe { std::ffi::CStr::from_ptr(a).to_str() };
        println!("{:?}", result)
    }

    #[test]
    fn test_account() {
        let from = wallet::restore_keypair_from_mnemonic_default("thought faint misery file cube cage agent flight gallery bundle thrive grant whip pig then purchase movie essence obey old cup loud until goose").unwrap();
        let pub_key = from.get_pk();
        let s = wallet::public_key_to_bech32(&pub_key);
        println!("{:?}", s);
    }

    #[test]
    fn test_all_account() {
        dotenv().ok();
        let accounts_result = tokio::runtime::Runtime::new().unwrap().block_on(async {
            let db_user = std::env::var("DBUSER").unwrap();
            let db_password = std::env::var("PASSWORD").unwrap();
            let db_host = std::env::var("HOST").unwrap();
            let db_port = std::env::var("PORT").unwrap();
            let db_name = std::env::var("DBName").unwrap();

            let uri = format!(
                "postgres://{}:{}@{}:{}/{}",
                db_user, db_password, db_host, db_port, db_name
            );

            let pg_pool = PgPoolOptions::new().connect(&uri).await.unwrap();
            Robot::all_accounts(&pg_pool)
                .await
                .unwrap_or_else(|_| vec![])
        });
        assert_eq!(200, accounts_result.len())
    }

    #[test]
    fn test_get_send_robot_batch_tx() {
        dotenv().ok();
        let mut from = std::env::var("CENTEREFROM").unwrap();
        let mut url = String::from("https://prod-testnet.prod.findora.org:8668");
        let a = get_send_robot_batch_tx(
            from.as_mut_ptr(),
            from.len() as u32,
            url.as_mut_ptr(),
            url.len() as u32,
        );
        let result = unsafe { std::ffi::CStr::from_ptr(a).to_str() };
        println!("{:?}", result)
    }

    #[test]
    fn test_get_user_fra_balance() {
        dotenv().ok();
        let mut from = std::env::var("CENTEREFROM").unwrap();
        let mut url = String::from("https://prod-testnet.prod.findora.org:8668");
        let a = get_user_fra_balance(
            from.as_mut_ptr(),
            from.len() as u32,
            url.as_mut_ptr(),
            url.len() as u32,
        );
        println!("{:?}", a)
    }
}
