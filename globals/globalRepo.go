package globals

import (
	"database/sql"
	"errors"
)

type sqlRepository struct {
	Conn *sql.DB
}

type H2hConfig struct {
	ID        int
	BaseURL   string
	PathURL   string
	MethodApi string
	Notes     string
}

func NewGlobalRepository(Conn *sql.DB) GlobalRepository {
	return &sqlRepository{Conn}
}

func (config *sqlRepository) RetrieveH2hConfigAPIEmail(id int) (*H2hConfig, error) {
	var data H2hConfig
	query := `SELECT base_url, path_url, method_api,notes FROM h2h_configs WHERE id = ? and (deleted_at = 0 OR deleted_at IS NULL)`
	rows, errQuery := config.Conn.Prepare(query)
	if errQuery != nil {
		return nil, errors.New("RetrieveH2hConfigAPIEmail errQuery = " + errQuery.Error())
	}
	defer rows.Close()

	errScan := rows.QueryRow(id).Scan(&data.BaseURL, &data.PathURL, &data.MethodApi, &data.Notes)
	if errScan != nil {

		if errScan == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.New("RetrieveH2hConfigAPIEmail.errScan = " + errScan.Error())
	}
	return &data, nil
}
