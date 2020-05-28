create table if not exists table_b
(
    id          serial not null
        constraint table_b_pk
            primary key,
    uniq_number integer,
    data        jsonb
);

alter table table_b
    owner to postgres;

create unique index table_b_uniq_number_uindex
    on table_b (uniq_number);

