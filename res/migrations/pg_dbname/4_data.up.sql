INSERT INTO public.table_a (id, text_uniq, number, not_null_number)
VALUES (1, 'a', null, 1);
INSERT INTO public.table_a (id, text_uniq, number, not_null_number)
VALUES (2, 'b', 1, 2);
INSERT INTO public.table_a (id, text_uniq, number, not_null_number)
VALUES (3, 'c', 2, 3);
INSERT INTO public.table_b (id, uniq_number, data)
VALUES (1, 1, '{}');
INSERT INTO public.table_b (id, uniq_number, data)
VALUES (2, 2, null);
INSERT INTO public.table_a_table_b (id, table_a_id, table_b_id)
VALUES (1, 1, 1);
INSERT INTO public.table_a_table_b (id, table_a_id, table_b_id)
VALUES (2, 1, 2);
INSERT INTO public.table_a_table_b (id, table_a_id, table_b_id)
VALUES (3, 2, 2);