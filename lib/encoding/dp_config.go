package libdrynxencoding

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DPLocalConfig 被 DP 进程在启动时设置
type DPLocalConfig struct {
	Database struct {
		Driver string `toml:"driver"` // e.g. "mysql", "postgres", "duckdb"
		DSN    string `toml:"dsn"`    // 推荐只读账号
	} `toml:"database"`
}

var dpLocalCfg *DPLocalConfig

// 由 DP 进程在 main 中调用注入
func SetDPLocalConfig(cfg *DPLocalConfig) {
	dpLocalCfg = cfg
}

// —— 下面是修改后的数据获取函数 ——
// 去掉 DptoPath 与 dpname（不再从外部拿 DSN），只接收逻辑表名与 SQL

// func GetDataFromDataProviderV2(tableLogical, sqlSurvey, opName string) ([]int64, error) {
// 	if dpLocalCfg == nil {
// 		return nil, fmt.Errorf("dp local config not loaded")
// 	}

// 	s := strings.TrimSpace(strings.ToUpper(sqlSurvey))
// 	if !strings.HasPrefix(s, "SELECT") {
// 		return nil, fmt.Errorf("only SELECT is allowed")
// 	}

// 	// 强制替换 FROM 后的表名，防止前端绕过（这里直接使用逻辑表名；
// 	// 若你想做“逻辑表→物理表”映射，可加一个 map）
// 	reFrom := regexp.MustCompile(`(?i)FROM\s+[\w\.\` + "`" + `"]+`)
// 	query := reFrom.ReplaceAllString(sqlSurvey, fmt.Sprintf("FROM `%s`", tableLogical))

// 	db, err := sql.Open(dpLocalCfg.Database.Driver, dpLocalCfg.Database.DSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("open db: %w", err)
// 	}
// 	defer db.Close()

// 	switch strings.ToLower(opName) {
// 	case "count", "sum", "mean", "avg", "max", "min":
// 		row := db.QueryRow(query)
// 		var v sql.NullFloat64
// 		if err := row.Scan(&v); err != nil {
// 			return nil, fmt.Errorf("scan agg: %w", err)
// 		}
// 		if !v.Valid {
// 			return []int64{0}, nil
// 		}
// 		return []int64{int64(math.Round(v.Float64))}, nil

// 	default:

// 		rows, err := db.Query(query)
// 		if err != nil {
// 			return nil, fmt.Errorf("query: %w", err)
// 		}
// 		defer rows.Close()

// 		out := make([]int64, 0, 1024)
// 		for rows.Next() {
// 			var p sql.NullFloat64
// 			if err := rows.Scan(&p); err != nil {
// 				return nil, fmt.Errorf("scan row: %w", err)
// 			}
// 			if p.Valid {
// 				out = append(out, int64(math.Round(p.Float64)))
// 			}
// 		}
// 		if err := rows.Err(); err != nil {
// 			return nil, err
// 		}
// 		return out, nil
// 	}
// }

// 依赖：import (
//   "database/sql"
//   "errors"
//   "fmt"
//   "math"
//   "strconv"
//   "strings"
// )

func GetDataFromDataProviderV3(sqlSurvey, opName string) ([]int64, error) {
	if dpLocalCfg == nil {
		return nil, fmt.Errorf("dp local config not loaded")
	}

	// 只允许 SELECT（避免注入/误用）
	s := strings.TrimSpace(sqlSurvey)
	if s == "" {
		return nil, errors.New("empty sql")
	}
	// 去掉末尾分号，禁止多语句（多语句也应在 DSN 上禁用）
	if strings.HasSuffix(s, ";") {
		s = strings.TrimRight(s, "; \t\r\n")
	}
	if !strings.HasPrefix(strings.ToUpper(strings.TrimLeft(s, " \t\r\n")), "SELECT") {
		return nil, fmt.Errorf("only SELECT is allowed")
	}

	db, err := sql.Open(dpLocalCfg.Database.Driver, dpLocalCfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	switch strings.ToLower(strings.TrimSpace(opName)) {
	case "count", "sum", "mean", "avg", "max", "min":
		row := db.QueryRow(s)
		val, err := scanNumericAsInt64(row)
		if err != nil {
			return nil, fmt.Errorf("scan agg: %w", err)
		}
		return []int64{val}, nil

	default:
		rows, err := db.Query(s)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		defer rows.Close()

		out := make([]int64, 0, 1024)
		for rows.Next() {
			val, err := scanRowFirstNumericAsInt64(rows)
			if err != nil {
				return nil, fmt.Errorf("scan row: %w", err)
			}
			out = append(out, val)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}
}

// ---- helpers ----

// 把单行单值（聚合）扫描为 int64（支持 int/float/decimal/bytes/string）
func scanNumericAsInt64(row *sql.Row) (int64, error) {
	var any interface{}
	if err := row.Scan(&any); err != nil {
		return 0, err
	}
	return coerceToInt64(any)
}

// 把每行第一列扫描为 int64（非聚合）
func scanRowFirstNumericAsInt64(rows *sql.Rows) (int64, error) {
	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}
	if len(cols) < 1 {
		return 0, errors.New("no columns in result")
	}
	var any interface{}
	if err := rows.Scan(&any); err != nil {
		return 0, err
	}
	return coerceToInt64(any)
}

// 尝试把多种数据库返回类型转成 int64（浮点取四舍五入）
func coerceToInt64(v interface{}) (int64, error) {
	switch t := v.(type) {
	case int64:
		return t, nil
	case int32:
		return int64(t), nil
	case int:
		return int64(t), nil
	case float64:
		return int64(math.Round(t)), nil
	case float32:
		return int64(math.Round(float64(t))), nil
	case []byte:
		// 可能是 DECIMAL/NUMERIC 或文本
		str := strings.TrimSpace(string(t))
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, nil
		}
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return int64(math.Round(f)), nil
		}
		return 0, fmt.Errorf("cannot parse numeric from bytes: %q", str)
	case string:
		str := strings.TrimSpace(t)
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, nil
		}
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return int64(math.Round(f)), nil
		}
		return 0, fmt.Errorf("cannot parse numeric from string: %q", str)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
