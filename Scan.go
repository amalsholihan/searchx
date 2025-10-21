package searchx

import (
	"gorm.io/gorm"
)

func (ks *Searchx) ScanOneToMap(tx *gorm.DB, dest *map[string]any) *Searchx {
	rows, err := tx.Limit(1).Rows()
	if err != nil {
		ks.Err = err
		return ks
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		ks.Err = err
		return ks
	}

	if !rows.Next() {
		// Tidak ada hasil, tetap kembalikan map kosong
		*dest = nil
		if err := rows.Err(); err != nil {
			ks.Err = err
		}
		return ks
	}

	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range cols {
		ptrs[i] = &values[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		ks.Err = err
		return ks
	}

	row := make(map[string]any, len(cols))
	for i, c := range cols {
		switch val := values[i].(type) {
		case []byte:
			row[c] = string(val)
		default:
			row[c] = val
		}
	}

	*dest = row

	if err := rows.Err(); err != nil {
		ks.Err = err
	}

	return ks
}

func (ks *Searchx) ScanAllToMap(tx *gorm.DB, dest *[]map[string]any) *Searchx {
	rows, err := tx.Rows()
	if err != nil {
		ks.Err = err
		return ks
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		ks.Err = err
		return ks
	}

	results := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range cols {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			ks.Err = err
			return ks
		}

		row := make(map[string]any, len(cols))
		for i, c := range cols {
			v := values[i]
			switch val := v.(type) {
			case []byte:
				row[c] = string(val)
			default:
				row[c] = val
			}
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		ks.Err = err
		return ks
	}

	*dest = results
	return nil
}
