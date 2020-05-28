create table if not exists table_a
(
    id              serial  not null
        constraint table_a_pk
            primary key,
    text_uniq       varchar,
    number          integer,
    not_null_number integer not null
);

alter table table_a
    owner to postgres;

create unique index table_a_text_uniq_uindex
    on table_a (text_uniq);

