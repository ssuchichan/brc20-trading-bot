extern crate dotenv;
use serde::Serialize;
use sqlx::Error;
use sqlx::PgPool;

#[derive(Serialize)]
pub struct Robot {
    mnemonic: String,
    account: String,
    create_time: i64,
    update_time: i64,
}

impl Robot {
    pub async fn all_accounts(pool: &PgPool) -> Result<Vec<String>, Error> {
        let result_list: Vec<(String,)> = sqlx::query_as("SELECT account FROM robot_list")
            .fetch_all(pool)
            .await?;
        let mut accounts_list: Vec<String> =
            result_list.into_iter().map(|(account,)| account).collect();

        let result_buy: Vec<(String,)> = sqlx::query_as("SELECT account FROM robot_buy")
            .fetch_all(pool)
            .await?;
        let accounts_buy: Vec<String> = result_buy.into_iter().map(|(account,)| account).collect();

        accounts_list.extend(accounts_buy);

        Ok(accounts_list)
    }
}

#[cfg(test)]
mod tests {
    use sqlx::postgres::PgPoolOptions;

    use super::Robot;
    use dotenv::dotenv;

    #[tokio::test]
    async fn test_all_accounts() {
        dotenv().ok();
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

        match Robot::all_accounts(&pg_pool).await {
            Ok(accounts) => {
                assert_eq!(200, accounts.len());
            }
            Err(e) => {
                eprintln!("{}", e)
            }
        }
    }
}
