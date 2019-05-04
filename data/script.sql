create table account_status
(
    account_status_id text not null,
    description       text not null
);

create unique index account_status_account_status_id_uindex
    on account_status (account_status_id);

create table balance
(
    private_id      text not null
        constraint balance_pk
            primary key,
    current_balance real not null,
    fee             real not null,
    update_counter  int  not null,
    currency_type   text
);

/* Account Table */
create table account
(
    public_id         text not null
        constraint account_pk
            primary key,
    private_id        text not null
        constraint account_balance__fk
            references balance,
    IPv4              text not null,
    email             text not null,
    registration_date text not null,
    account_status    text not null
);

create unique index account_public_id_uindex
    on account (public_id);

create unique index account_system_id_uindex
    on account (private_id);

create unique index balance_public_id_uindex
    on balance (private_id);

create table "transaction"
(
    transaction_id   text not null
        constraint transaction_pk
            primary key,
    private_id       text not null,
    contra_id        text not null,
    amount           real not null,
    currency_type    text not null,
    transaction_type text
);

create table currency_type
(
    symbol   text not null
        constraint currency_type_transaction_currency_type_fk
            references "transaction" (currency_type),
    currency text not null
);

create unique index currency_type_symbol_uindex
    on currency_type (symbol);

create unique index transaction_transaction_id_uindex
    on "transaction" (transaction_id);

create table transaction_type
(
    transaction_type_id text not null
        constraint transaction_type_fk
            references "transaction" (transaction_type),
    description         text not null
);


