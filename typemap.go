package dgw

const typeMap = `
[string]
db_types = ["character", "character varying", "text", "money"]
notnull_go_type = "string"
notnull_nil_value= "\"\""
nullable_go_type = "sql.NullString"
nullable_nil_value= "\"\""

[time]
db_types = [
    "time with time zone", "time without time zone",
    "timestamp without time zone", "timestamp with time zone", "date"
]
notnull_go_type = "time.Time"
notnull_nil_value= "0"
nullable_go_type = "*time.Time"
nullable_nil_value= "0"

[bool]
db_types = ["boolean"]
notnull_go_type = "boole"
notnull_nil_value= "false"
nullable_go_type = "boole"
nullable_nil_value= "false"

[smallint]
db_types = ["smallint"]
notnull_go_type = "int16"
notnull_nil_value= "0"
nullable_go_type = "sql.NullInt64"
nullable_nil_value= "sql.NullInt64{}"

[integer]
db_types = ["integer"]
notnull_go_type = "int"
notnull_nil_value= "0"
nullable_go_type = "sql.NullInt64"
nullable_nil_value= "sql.NullInt64{}"

[bigint]
db_types = ["bigint"]
notnull_go_type = "int64"
notnull_nil_value= "0"
nullable_go_type = "sql.NullInt64"
nullable_nil_value= "sql.NullInt64{}"

[smallserial]
db_types = ["smallserial"]
notnull_go_type = "uint16"
notnull_nil_value= "0"
nullable_go_type = "sql.NullInt64"
nullable_nil_value= "sql.NullInt64{}"

[serial]
db_types = ["serial"]
notnull_go_type = "uint32"
notnull_nil_value= "0"
nullable_go_type = "sql.NullInt64"
nullable_nil_value= "sql.NullInt64{}"

[real]
db_types = ["real"]
notnull_go_type = "float32"
notnull_nil_value= "0.0"
nullable_go_type = "sql.NullFloat64"
nullable_nil_value= "sql.NullFloat64{}"

[numeric]
db_types = ["numeric", "double precision"]
notnull_go_type = "float64"
notnull_nil_value= "0.0"
nullable_go_type = "sql.NullFloat64"
nullable_nil_value= "sql.NullFloat64{}"

[bytea]
db_types = ["bytea"]
notnull_go_type = "byte"
notnull_nil_value= "\"\""
nullable_go_type = "byte"
nullable_nil_value= "\"\""

[interval]
db_types = ["interval"]
notnull_go_type = "time.Duration"
notnull_nil_value= "0"
nullable_go_type = "*time.Duration"
nullable_nil_value= "0"

[default]
db_types = ["*"]
notnull_go_type = "interface{}"
notnull_nil_value= "nil"
nullable_go_type = "interface{}"
`
