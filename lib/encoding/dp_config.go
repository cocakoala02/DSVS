package libdrynxencoding

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// DPLocalConfig 被 DP 进程在启动时设置
type DPLocalConfig struct {
	Database struct {
		Driver string `toml:"driver"` // e.g. "mysql", "postgres", "duckdb"
		DSN    string `toml:"dsn"`    // 推荐只读账号
	} `toml:"database"`

	// 新增：哪些字段需要定点缩放
}

var dpLocalCfg *DPLocalConfig

// 由 DP 进程在 main 中调用注入
func SetDPLocalConfig(cfg *DPLocalConfig) {
	dpLocalCfg = cfg
}

// func GetDataFromDataProviderV3(sqlSurvey, opName string) ([]int64, error) {
// 	if dpLocalCfg == nil {
// 		return nil, fmt.Errorf("dp local config not loaded")
// 	}

// 	// 只允许 SELECT（避免注入/误用）
// 	s := strings.TrimSpace(sqlSurvey)
// 	if s == "" {
// 		return nil, errors.New("empty sql")
// 	}
// 	// 去掉末尾分号，禁止多语句（多语句也应在 DSN 上禁用）
// 	if strings.HasSuffix(s, ";") {
// 		s = strings.TrimRight(s, "; \t\r\n")
// 	}
// 	if !strings.HasPrefix(strings.ToUpper(strings.TrimLeft(s, " \t\r\n")), "SELECT") {
// 		return nil, fmt.Errorf("only SELECT is allowed")
// 	}

// 	db, err := sql.Open(dpLocalCfg.Database.Driver, dpLocalCfg.Database.DSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("open db: %w", err)
// 	}
// 	defer db.Close()

// 	switch strings.ToLower(strings.TrimSpace(opName)) {
// 	case "count", "sum", "mean", "avg", "max", "min":
// 		row := db.QueryRow(s)
// 		val, err := scanNumericAsInt64(row)
// 		if err != nil {
// 			return nil, fmt.Errorf("scan agg: %w", err)
// 		}
// 		return []int64{val}, nil

// 	default:
// 		rows, err := db.Query(s)
// 		if err != nil {
// 			return nil, fmt.Errorf("query: %w", err)
// 		}
// 		defer rows.Close()

// 		out := make([]int64, 0, 1024)
// 		for rows.Next() {
// 			val, err := scanRowFirstNumericAsInt64(rows)
// 			if err != nil {
// 				return nil, fmt.Errorf("scan row: %w", err)
// 			}
// 			out = append(out, val)
// 		}
// 		if err := rows.Err(); err != nil {
// 			return nil, err
// 		}
// 		return out, nil
// 	}
// }

// libdrynxencoding/dp_config.go
// func GetDataFromDataProviderV4(sqlSurvey, opName string, fixedscale int64) ([]int64, error) {
// 	if dpLocalCfg == nil {
// 		return nil, fmt.Errorf("dp local config not loaded")
// 	}
// 	s := strings.TrimSpace(sqlSurvey)
// 	if s == "" {
// 		return nil, errors.New("empty sql")
// 	}
// 	if strings.HasSuffix(s, ";") {
// 		s = strings.TrimRight(s, "; \t\r\n")
// 	}
// 	if !strings.HasPrefix(strings.ToUpper(strings.TrimLeft(s, " \t\r\n")), "SELECT") {
// 		return nil, fmt.Errorf("only SELECT is allowed")
// 	}

// 	db, err := sql.Open(dpLocalCfg.Database.Driver, dpLocalCfg.Database.DSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("open db: %w", err)
// 	}
// 	defer db.Close()

// 	switch strings.ToLower(strings.TrimSpace(opName)) {
// 	case "count":
// 		// 直接执行你提供的 COUNT 查询，并扫描第一列
// 		row := db.QueryRow(s)
// 		v, err := scanNumericAsInt64(row)
// 		if err != nil {
// 			return nil, fmt.Errorf("scan count: %w", err)
// 		}
// 		return []int64{v}, nil

// 		// case "sum":
// 		// 	// 直接执行 SUM 查询，并扫描第一列
// 		// 	row := db.QueryRow(s)
// 		// 	v, err := scanNumericAsInt64(row)
// 		// 	if err != nil {
// 		// 		return nil, fmt.Errorf("scan sum: %w", err)
// 		// 	}
// 		// 	return []int64{v}, nil

// 		// case "mean", "avg":
// 		// 	// 把 `SELECT AVG(expr) FROM ...` 自动重写为 `SELECT COALESCE(SUM(expr),0), COUNT(expr) FROM ...`
// 		// 	reAvg := regexp.MustCompile(`(?is)^\s*SELECT\s+AVG\s*\(\s*(?P<expr>.+?)\s*\)\s+FROM\s+(?P<rest>.+)$`)
// 		// 	m := reAvg.FindStringSubmatch(s)
// 		// 	if m == nil {
// 		// 		return nil, fmt.Errorf("AVG rewrite failed: expected `SELECT AVG(expr) FROM ...`, got: %q", s)
// 		// 	}
// 		// 	expr := strings.TrimSpace(m[reAvg.SubexpIndex("expr")])
// 		// 	rest := strings.TrimSpace(m[reAvg.SubexpIndex("rest")])

// 		// 	q := fmt.Sprintf("SELECT COALESCE(SUM(%s),0), COUNT(%s) FROM %s", expr, expr, rest)
// 		// 	row := db.QueryRow(q)

// 		// 	var a, b interface{}
// 		// 	if err := row.Scan(&a, &b); err != nil {
// 		// 		return nil, fmt.Errorf("scan mean(sum,count): %w", err)
// 		// 	}
// 		// 	sum, err := coerceToInt64(a)
// 		// 	if err != nil {
// 		// 		return nil, fmt.Errorf("coerce sum: %w", err)
// 		// 	}
// 		// 	cnt, err := coerceToInt64(b)
// 		// 	if err != nil {
// 		// 		return nil, fmt.Errorf("coerce count: %w", err)
// 		// 	}
// 		// 	return []int64{sum, cnt}, nil

// 		// case "min":
// 		// 	// 仍包一层，对单列结果取最小值
// 		// 	q := fmt.Sprintf("SELECT COALESCE(MIN(__v),0) FROM ( %s ) AS _q(__v)", s)
// 		// 	row := db.QueryRow(q)
// 		// 	v, err := scanNumericAsInt64(row)
// 		// 	if err != nil {
// 		// 		return nil, fmt.Errorf("scan min: %w", err)
// 		// 	}
// 		// 	return []int64{v}, nil

// 		// case "max":
// 		// 	// 仍包一层，对单列结果取最大值
// 		// 	q := fmt.Sprintf("SELECT COALESCE(MAX(__v),0) FROM ( %s ) AS _q(__v)", s)
// 		// 	row := db.QueryRow(q)
// 		// 	v, err := scanNumericAsInt64(row)
// 		// 	if err != nil {
// 		// 		return nil, fmt.Errorf("scan max: %w", err)
// 		// 	}
// 		// 	return []int64{v}, nil

// 	case "sum":
// 		row := db.QueryRow(s)
// 		var any interface{}
// 		if err := row.Scan(&any); err != nil {
// 			return nil, fmt.Errorf("scan sum: %w", err)
// 		}
// 		// fmt.Printf("ssssssssssssssssssssssssssssssssssssssssssss")
// 		log.LLvl2(any)
// 		vs, err := coerceToScaledInt64(any, fixedscale)
// 		if err != nil {
// 			return nil, fmt.Errorf("scale sum: %w", err)
// 		}
// 		log.LLvl2(vs)
// 		// fmt.Printf("ssssssssssssssssssssssssssssssssssssssssssss")

// 		return []int64{vs}, nil

// 	case "mean", "avg":
// 		reAvg := regexp.MustCompile(`(?is)^\s*SELECT\s+AVG\s*\(\s*(?P<expr>.+?)\s*\)\s+FROM\s+(?P<rest>.+)$`)
// 		m := reAvg.FindStringSubmatch(s)
// 		if m == nil {
// 			return nil, fmt.Errorf("AVG rewrite failed: expected `SELECT AVG(expr) FROM ...`")
// 		}
// 		expr := strings.TrimSpace(m[reAvg.SubexpIndex("expr")])
// 		rest := strings.TrimSpace(m[reAvg.SubexpIndex("rest")])

// 		q := fmt.Sprintf("SELECT COALESCE(SUM(%s),0), COUNT(%s) FROM %s", expr, expr, rest)
// 		row := db.QueryRow(q)

// 		var a, b interface{}
// 		if err := row.Scan(&a, &b); err != nil {
// 			return nil, fmt.Errorf("scan mean(sum,count): %w", err)
// 		}
// 		sumScaled, err := coerceToScaledInt64(a, fixedscale)
// 		if err != nil {
// 			return nil, fmt.Errorf("scale sum: %w", err)
// 		}
// 		cnt, err := coerceToInt64(b) // 这里还是整数
// 		if err != nil {
// 			return nil, fmt.Errorf("coerce count: %w", err)
// 		}
// 		return []int64{sumScaled, cnt}, nil

// 	case "min":
// 		q := fmt.Sprintf("SELECT COALESCE(MIN(__v),0) FROM ( %s ) AS _q(__v)", s)
// 		row := db.QueryRow(q)
// 		var any interface{}
// 		if err := row.Scan(&any); err != nil {
// 			return nil, fmt.Errorf("scan min: %w", err)
// 		}
// 		vScaled, err := coerceToScaledInt64(any, fixedscale)
// 		if err != nil {
// 			return nil, fmt.Errorf("scale min: %w", err)
// 		}
// 		return []int64{vScaled}, nil

// 	case "max":
// 		q := fmt.Sprintf("SELECT COALESCE(MAX(__v),0) FROM ( %s ) AS _q(__v)", s)
// 		row := db.QueryRow(q)
// 		var any interface{}
// 		if err := row.Scan(&any); err != nil {
// 			return nil, fmt.Errorf("scan max: %w", err)
// 		}
// 		vScaled, err := coerceToScaledInt64(any, fixedscale)
// 		if err != nil {
// 			return nil, fmt.Errorf("scale max: %w", err)
// 		}
// 		return []int64{vScaled}, nil

// 	default:
// 		return nil, fmt.Errorf("unsupported op %q", opName)
// 	}
// }

// lib/encoding/dp_config.go
// lib/encoding/dp_config.go
// func GetDataFromDataProviderV4(sqlSurvey, opName string, scaleS int64) ([]int64, error) {
// 	if dpLocalCfg == nil {
// 		return nil, fmt.Errorf("dp local config not loaded")
// 	}
// 	s := strings.TrimSpace(sqlSurvey)
// 	if s == "" {
// 		return nil, errors.New("empty sql")
// 	}
// 	if strings.HasSuffix(s, ";") {
// 		s = strings.TrimRight(s, "; \t\r\n")
// 	}
// 	if !strings.HasPrefix(strings.ToUpper(strings.TrimLeft(s, " \t\r\n")), "SELECT") {
// 		return nil, fmt.Errorf("only SELECT is allowed")
// 	}

// 	// 只根据 DP 的 float_columns + SQL 判定是否缩放
// 	scaled, _ := decideScaledBySQL(s, opName)

// 	db, err := sql.Open(dpLocalCfg.Database.Driver, dpLocalCfg.Database.DSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("open db: %w", err)
// 	}
// 	defer db.Close()

// 	switch strings.ToLower(strings.TrimSpace(opName)) {
// 	case "count":
// 		row := db.QueryRow(s)
// 		v, err := scanNumericAsInt64(row)
// 		if err != nil {
// 			return nil, fmt.Errorf("scan count: %w", err)
// 		}
// 		return []int64{v}, nil

// 	case "sum":
// 		row := db.QueryRow(s)
// 		var any interface{}
// 		if err := row.Scan(&any); err != nil {
// 			return nil, fmt.Errorf("scan sum: %w", err)
// 		}
// 		v, err := coerceToInt64MaybeScaled(any, scaled, scaleS)
// 		if err != nil {
// 			return nil, fmt.Errorf("sum scale: %w", err)
// 		}
// 		return []int64{v}, nil

// 	case "mean", "avg":
// 		reAvg := regexp.MustCompile(`(?is)^\s*SELECT\s+AVG\s*\(\s*(?P<expr>.+?)\s*\)\s+FROM\s+(?P<rest>.+)$`)
// 		m := reAvg.FindStringSubmatch(s)
// 		if m == nil {
// 			return nil, fmt.Errorf("AVG rewrite failed: expected `SELECT AVG(expr) FROM ...`, got: %q", s)
// 		}
// 		expr := strings.TrimSpace(m[reAvg.SubexpIndex("expr")])
// 		rest := strings.TrimSpace(m[reAvg.SubexpIndex("rest")])
// 		q := fmt.Sprintf("SELECT COALESCE(SUM(%s),0), COUNT(%s) FROM %s", expr, expr, rest)

// 		row := db.QueryRow(q)
// 		var a, b interface{}
// 		if err := row.Scan(&a, &b); err != nil {
// 			return nil, fmt.Errorf("scan mean(sum,count): %w", err)
// 		}
// 		sumScaled, err := coerceToInt64MaybeScaled(a, scaled, scaleS) // 只缩放 sum
// 		if err != nil {
// 			return nil, fmt.Errorf("coerce sum: %w", err)
// 		}
// 		cnt, err := coerceToInt64(b)
// 		if err != nil {
// 			return nil, fmt.Errorf("coerce count: %w", err)
// 		}
// 		return []int64{sumScaled, cnt}, nil

// 	case "min":
// 		q := fmt.Sprintf("SELECT COALESCE(MIN(__v),0) FROM ( %s ) AS _q(__v)", s)
// 		row := db.QueryRow(q)
// 		var any interface{}
// 		if err := row.Scan(&any); err != nil {
// 			return nil, fmt.Errorf("scan min: %w", err)
// 		}
// 		v, err := coerceToInt64MaybeScaled(any, scaled, scaleS)
// 		if err != nil {
// 			return nil, fmt.Errorf("min scale: %w", err)
// 		}
// 		return []int64{v}, nil

// 	case "max":
// 		q := fmt.Sprintf("SELECT COALESCE(MAX(__v),0) FROM ( %s ) AS _q(__v)", s)
// 		row := db.QueryRow(q)
// 		var any interface{}
// 		if err := row.Scan(&any); err != nil {
// 			return nil, fmt.Errorf("scan max: %w", err)
// 		}
// 		v, err := coerceToInt64MaybeScaled(any, scaled, scaleS)
// 		if err != nil {
// 			return nil, fmt.Errorf("max scale: %w", err)
// 		}
// 		return []int64{v}, nil

// 	default:
// 		return nil, fmt.Errorf("unsupported op %q", opName)
// 	}
// }
// 删掉：FloatColumns []string `toml:"float_columns"`

// V5：由调用方（DP 上游）传入 floatColumns（来自客户端），DP 不再有本地浮点列来源
func GetDataFromDataProviderV5(sqlSurvey, opName string, scaleS int64, floatColumns []string) ([]int64, error) {
	if dpLocalCfg == nil {
		return nil, fmt.Errorf("dp local config not loaded")
	}
	s := strings.TrimSpace(sqlSurvey)
	if s == "" {
		return nil, errors.New("empty sql")
	}
	if strings.HasSuffix(s, ";") {
		s = strings.TrimRight(s, "; \t\r\n")
	}
	if !strings.HasPrefix(strings.ToUpper(strings.TrimLeft(s, " \t\r\n")), "SELECT") {
		return nil, fmt.Errorf("only SELECT is allowed")
	}

	// 仅用客户端提供的 floatColumns + SQL 判定是否缩放
	scaled := decideScaledBySQLOnList(s, floatColumns)

	db, err := sql.Open(dpLocalCfg.Database.Driver, dpLocalCfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	switch strings.ToLower(strings.TrimSpace(opName)) {
	case "count":
		row := db.QueryRow(s)
		v, err := scanNumericAsInt64(row)
		if err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		return []int64{v}, nil

	case "sum":
		row := db.QueryRow(s)
		var any interface{}
		if err := row.Scan(&any); err != nil {
			return nil, fmt.Errorf("scan sum: %w", err)
		}
		v, err := coerceToInt64MaybeScaled(any, scaled, scaleS)
		if err != nil {
			return nil, fmt.Errorf("sum scale: %w", err)
		}
		return []int64{v}, nil

	case "mean", "avg":
		reAvg := regexp.MustCompile(`(?is)^\s*SELECT\s+AVG\s*\(\s*(?P<expr>.+?)\s*\)\s+FROM\s+(?P<rest>.+)$`)
		m := reAvg.FindStringSubmatch(s)
		if m == nil {
			return nil, fmt.Errorf("AVG rewrite failed: expected `SELECT AVG(expr) FROM ...`, got: %q", s)
		}
		expr := strings.TrimSpace(m[reAvg.SubexpIndex("expr")])
		rest := strings.TrimSpace(m[reAvg.SubexpIndex("rest")])
		q := fmt.Sprintf("SELECT COALESCE(SUM(%s),0), COUNT(%s) FROM %s", expr, expr, rest)

		row := db.QueryRow(q)
		var a, b interface{}
		if err := row.Scan(&a, &b); err != nil {
			return nil, fmt.Errorf("scan mean(sum,count): %w", err)
		}
		sumScaled, err := coerceToInt64MaybeScaled(a, scaled, scaleS) // 仅 sum 缩放
		if err != nil {
			return nil, fmt.Errorf("coerce sum: %w", err)
		}
		cnt, err := coerceToInt64(b)
		if err != nil {
			return nil, fmt.Errorf("coerce count: %w", err)
		}
		return []int64{sumScaled, cnt}, nil

	case "min":
		q := fmt.Sprintf("SELECT COALESCE(MIN(__v),0) FROM ( %s ) AS _q(__v)", s)
		row := db.QueryRow(q)
		var any interface{}
		if err := row.Scan(&any); err != nil {
			return nil, fmt.Errorf("scan min: %w", err)
		}
		v, err := coerceToInt64MaybeScaled(any, scaled, scaleS)
		if err != nil {
			return nil, fmt.Errorf("min scale: %w", err)
		}
		return []int64{v}, nil

	case "max":
		q := fmt.Sprintf("SELECT COALESCE(MAX(__v),0) FROM ( %s ) AS _q(__v)", s)
		row := db.QueryRow(q)
		var any interface{}
		if err := row.Scan(&any); err != nil {
			return nil, fmt.Errorf("scan max: %w", err)
		}
		v, err := coerceToInt64MaybeScaled(any, scaled, scaleS)
		if err != nil {
			return nil, fmt.Errorf("max scale: %w", err)
		}
		return []int64{v}, nil

	default:
		return nil, fmt.Errorf("unsupported op %q", opName)
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
func coerceToInt64MaybeScaled(v interface{}, doScale bool, scaleS int64) (int64, error) {
	if !doScale || scaleS <= 1 {
		return coerceToInt64(v)
	}
	switch t := v.(type) {
	case int64, int32, int:
		i, _ := coerceToInt64(v)
		return i * scaleS, nil
	case float64:
		return int64(math.Round(t * float64(scaleS))), nil
	case float32:
		return int64(math.Round(float64(t) * float64(scaleS))), nil
	case []byte:
		s := strings.TrimSpace(string(t))
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(math.Round(f * float64(scaleS))), nil
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i * scaleS, nil
		}
		return 0, fmt.Errorf("cannot parse numeric from bytes: %q", s)
	case string:
		s := strings.TrimSpace(t)
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(math.Round(f * float64(scaleS))), nil
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i * scaleS, nil
		}
		return 0, fmt.Errorf("cannot parse numeric from string: %q", s)
	case nil:
		return 0, nil
	default:
		return coerceToInt64(v) // 兜底
	}
}

// // 仅使用 DP 的 toml（唯一来源）
// func isFloatColumn(expr string) bool {
// 	if dpLocalCfg == nil || len(dpLocalCfg.Database.FloatColumns) == 0 {
// 		return false
// 	}
// 	name := strings.ToLower(strings.TrimSpace(expr))
// 	if i := strings.LastIndex(name, "."); i >= 0 {
// 		name = name[i+1:]
// 	}
// 	for _, c := range dpLocalCfg.Database.FloatColumns {
// 		if strings.ToLower(strings.TrimSpace(c)) == name {
// 			return true
// 		}
// 	}
// 	return false
// }

// func decideScaledBySQL(sql, opName string) (bool, string) {
// 	s := strings.TrimSpace(sql)
// 	reAgg := regexp.MustCompile(`(?is)^\s*SELECT\s+(?P<func>SUM|AVG|MIN|MAX)\s*\(\s*(?P<expr>[a-zA-Z_][\w\.]*)\s*\)`)
// 	if m := reAgg.FindStringSubmatch(s); m != nil {
// 		expr := m[reAgg.SubexpIndex("expr")]
// 		return isFloatColumn(expr), expr
// 	}
// 	reFirst := regexp.MustCompile(`(?is)^\s*SELECT\s+(?P<expr>[a-zA-Z_][\w\.]*)\s+FROM\s+`)
// 	if m := reFirst.FindStringSubmatch(s); m != nil {
// 		expr := m[reFirst.SubexpIndex("expr")]
// 		return isFloatColumn(expr), expr
// 	}
// 	return false, ""
// }

// 根据 SQL 提取第一个目标表达式（SUM/AVG/MIN/MAX(expr) 或 SELECT expr FROM ...）
func extractFirstExpr(sql string) string {
	s := strings.TrimSpace(sql)
	reAgg := regexp.MustCompile(`(?is)^\s*SELECT\s+(SUM|AVG|MIN|MAX)\s*\(\s*([a-zA-Z_][\w\.]*)\s*\)`)
	if m := reAgg.FindStringSubmatch(s); m != nil {
		return m[2]
	}
	reFirst := regexp.MustCompile(`(?is)^\s*SELECT\s+([a-zA-Z_][\w\.]*)\s+FROM\s+`)
	if m := reFirst.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}

// 只用“传进来的” floatColumns 判定（客户端来源）
func decideScaledBySQLOnList(sql string, floatColumns []string) bool {
	expr := extractFirstExpr(sql)
	if expr == "" {
		return false
	}
	name := strings.ToLower(strings.TrimSpace(expr))
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	for _, c := range floatColumns {
		cc := strings.ToLower(strings.TrimSpace(c))
		// 支持 "col" 或 "tbl.col" 两种写法
		if cc == name || strings.HasSuffix(cc, "."+name) {
			return true
		}
	}
	return false
}
