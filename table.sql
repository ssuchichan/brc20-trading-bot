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

