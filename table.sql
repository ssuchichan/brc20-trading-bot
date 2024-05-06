CREATE SEQUENCE "balance_id_seq"
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

CREATE TABLE "balance" (
  "id" bigint NOT NULL DEFAULT nextval('balance_id_seq'::regclass),
  "address" varchar(64) NOT NULL,
  "ticker" varchar(45) NOT NULL,
  "overall_balance" Numeric NOT NULL,
  "create_time" int4 NOT NULL,
  "update_time" int4 NOT NULL,
  "height" int8 NOT NULL,
  CONSTRAINT "balance_pkey" PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "id_UNIQUE_balance" ON "balance" USING btree (
  "id" ASC
);
CREATE INDEX "search_index" ON "balance" USING btree (
  "ticker" ASC,
  "address" ASC
);

CREATE SEQUENCE "token_id_seq"
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

CREATE TABLE "token" (
  "id" bigint NOT NULL DEFAULT nextval('token_id_seq'::regclass),
  "ticker" varchar(45) NOT NULL,
  "dec" int4 NOT NULL,
  "max" Numeric NOT NULL,
  "lim" Numeric NOT NULL,
  "create_time" int4 NOT NULL,
  "update_time" int4 NOT NULL,
  "deploy_user" varchar(64) NOT NULL,
  CONSTRAINT "token_pkey" PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "id_UNIQUE_token" ON "token" USING btree (
  "id" ASC
);
CREATE UNIQUE INDEX "ticker_UNIQUE_token" on "token" USING btree (
  "ticker" ASC
);
CREATE INDEX "search_index_copy_1" ON "token" USING btree (
  "ticker" ASC
);

CREATE SEQUENCE "mint_record_id_seq"
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

CREATE TABLE "mint_record" (
  "id" bigint NOT NULL DEFAULT nextval('mint_record_id_seq'::regclass),
  "ticker" varchar(45) NOT NULL,
  "user" varchar(64) NOT NULL,
  "create_time" int4 NOT NULL,
  "amount" Numeric NOT NULL,
  "update_time" int4 NOT NULL,
  CONSTRAINT "mint_record_pkey" PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "id_UNIQUE_mint_record" ON "mint_record" USING btree (
  "id" ASC
);
CREATE INDEX "search" ON "mint_record" USING btree (
  "ticker" ASC,
  "user" ASC
);


CREATE SEQUENCE "list_record_id_seq"
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE "list_record" (
  "id" bigint NOT NULL DEFAULT nextval('list_record_id_seq'::regclass),
  "ticker" varchar(45) NOT NULL,
  "user" varchar(64) NOT NULL,
  "amount" Numeric NOT NULL,
  "price" Numeric NOT NULL,
  "state" integer NOT NULL DEFAULT 0,
  "to_user" varchar(64) DEFAULT '',
  "center_mnemonic" varchar(512) DEFAULT '',
  "create_time" int4 NOT NULL,
  "update_time" int4 NOT NULL,
  CONSTRAINT "list_record_pkey" PRIMARY KEY ("id")
);
COMMENT ON COLUMN list_record.state
    IS '0: 上架中 , 1: 取消, 2: 已完成, 3: 待上架';
CREATE UNIQUE INDEX "id_UNIQUE_list_record" ON "list_record" USING btree (
  "id" ASC
);

CREATE SEQUENCE "airdrop_record_id_seq"
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE "airdrop_record" (
  "id" bigint NOT NULL DEFAULT nextval('airdrop_record_id_seq'::regclass),
  "from_user" varchar(64) NOT NULL,
  "to_user" varchar(64) NOT NULL,
  "amount" Numeric NOT NULL,
  "create_time" int4 NOT NULL,
  "update_time" int4 NOT NULL,
  CONSTRAINT "airdrop_record_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "id_UNIQUE_airdrop_record" ON "airdrop_record" USING btree (
  "id" ASC
);

CREATE INDEX "search_airdrop" ON "airdrop_record" USING btree (
  "from_user" ASC,
  "to_user" ASC
);


CREATE SEQUENCE "robot_list_id_seq"
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE "robot_list" (
  "id" bigint NOT NULL DEFAULT nextval('robot_list_id_seq'::regclass),
  "account" varchar(64) NOT NULL,
  "private_key" varchar(48) NOT NULL,
  "public_key" varchar(48) NOT NULL,
  "mnemonic" varchar(512) NOT NULL,
  "create_time" int4 NOT NULL,
  "update_time" int4 NOT NULL,
  CONSTRAINT "robot_list_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "id_UNIQUE_robot_list" ON "robot_list" USING btree ("id" ASC);

CREATE SEQUENCE "robot_buy_id_seq"
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE "robot_buy" (
    "id" bigint NOT NULL DEFAULT nextval('robot_buy_id_seq'::regclass),
    "account" varchar(64) NOT NULL,
    "private_key" varchar(48) NOT NULL,
    "public_key" varchar(48) NOT NULL,
    "mnemonic" varchar(512) NOT NULL,
    "create_time" int4 NOT NULL,
    "update_time" int4 NOT NULL,
    CONSTRAINT "robot_buy_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "id_UNIQUE_robot_buy" ON "robot_buy" USING btree ("id" ASC);

