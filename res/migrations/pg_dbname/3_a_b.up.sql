create table if not exists table_a_table_b
(
    id         serial  not null
        constraint table_a_table_b_pk
            primary key,
    table_a_id integer not null
        constraint table_a_table_b_table_a_id_fk
            references table_a
            on update restrict on delete restrict,
    table_b_id integer not null
        constraint table_a_table_b_table_b_id_fk
            references table_b
            on update restrict on delete restrict
);

alter table table_a_table_b
    owner to postgres;

