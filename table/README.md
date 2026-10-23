# PostgreSQL Menu Table Structs

These Go files match the table definitions in `menu_schema.sql`, under package `database`.

## Collection-to-table mapping

- `MenuStruct` metadata -> `ReceptionMenuStruct` (`reception_menu`)
- `Menu` -> `ReceptionMenuItemStruct` (`reception_menu_item`)
- `Menu.Items` -> `ReceptionMenuItemContentStruct`
- `Menu.PaidOptions` (`[]ItemOption`) -> `ReceptionMenuPaidOptionStruct`
- `Menu.FreeOptions` (`[][]int`) -> `ReceptionMenuFreeOptionGroupStruct` + `ReceptionMenuFreeOptionItemStruct`
- `Menu.FreeMultiOptions` -> `ReceptionMenuFreeMultiOptionStruct`
- `ItemDetail` -> `ReceptionMenuItemDetailStruct`
- `ItemDetail.Choices` (`[][]string`) -> `ReceptionMenuItemDetailChoiceGroupStruct` + `ReceptionMenuItemDetailChoiceStruct`

All `BIGINT` columns, including identity IDs and application `menu_id` / `item_id`, use Go `int64`.
PostgreSQL `INTEGER` columns use Go `int`.
`sort_order` preserves original slice order.

The `db` struct tags document SQL column names; the standard `database/sql` package does not scan fields by `db` tag automatically.
Where nullable SQL fields are scanned into plain `string` / `int` values, use `COALESCE` in the SELECT statement (as in the existing ReceptionBook query).
For PostgreSQL arrays in query arguments or scans, use `pq.Array` when needed. These normalized tables store the old nested slices as rows, not PostgreSQL arrays.
