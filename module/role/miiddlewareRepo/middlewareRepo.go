package miiddlewareRepo

import (
	"database/sql"
	"errors"

	"github.com/YugaAdiIrawan/model/role"
)

type MiddlewareRepository interface {
	RetrieveUserByUserID(userID int) (*role.ProfileUser, error)
	RetrieveUserRoleAllowanceIsExist(roleID int, method, route string) (bool, error)
	RetrieveRouteAlreadyRegistered(method, route string) (bool, error)
}

type sqlRepository struct {
	Conn *sql.DB
}

// NewMiddlewareRepository - will create object that represent
func NewMiddlewareRepository(Conn *sql.DB) MiddlewareRepository {
	return &sqlRepository{Conn}
}

func (config *sqlRepository) RetrieveUserByUserID(userID int) (*role.ProfileUser, error) {
	query := `SELECT full_name,gender,avatar,role_id , is_blocked,role_name
				FROM vw_users WHERE user_id = ?`

	stat, errPrepare := config.Conn.Prepare(query)
	if errPrepare != nil {
		return nil, errors.New("RetrieveUserByUserID.errPrepare : " + errPrepare.Error())
	}
	defer stat.Close()

	var data role.ProfileUser
	errQueryRow := stat.QueryRow(userID).Scan(&data.FullName, &data.Gender, &data.Avatar, &data.RoleID, &data.IsBlocked, &data.RoleName)
	if errQueryRow != nil {
		if errQueryRow == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.New("RetrieveUserByUserID.errQueryRow : " + errQueryRow.Error())
	}

	return &data, nil
}

func (config *sqlRepository) RetrieveRouteAlreadyRegistered(method, route string) (bool, error) {
	query := `SELECT EXISTS(select 1 FROM api_rests WHERE api_rests.method = ? AND api_rests.rest_url = ? and api_rests.deleted_at = 0)`

	// Prepare the query
	stat, errPrepare := config.Conn.Prepare(query)
	if errPrepare != nil {
		return false, errors.New("RetrieveRouteAlreadyRegistered.errPrepare : " + errPrepare.Error())
	}
	defer stat.Close()

	// Execute the query
	var exists bool
	errQueryRow := stat.QueryRow(method, route).Scan(&exists)
	if errQueryRow != nil {
		if errQueryRow == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.New("RetrieveRouteAlreadyRegistered.errQueryRow : " + errQueryRow.Error())
	}

	return exists, nil
}

func (config *sqlRepository) RetrieveUserRoleAllowanceIsExist(roleID int, method, route string) (bool, error) {
	query := `SELECT EXISTS(select 1 FROM api_rests LEFT JOIN role_apis ON api_rests.id = role_apis.api_rest_id WHERE role_apis.role_id = ? AND api_rests.method = ? AND api_rests.rest_url = ? and role_apis.deleted_at = 0)`

	// Prepare the query
	stat, errPrepare := config.Conn.Prepare(query)
	if errPrepare != nil {
		return false, errors.New("RetrieveUserRoleAllowanceIsExist.errPrepare : " + errPrepare.Error())
	}
	defer stat.Close()

	// Execute the query
	var exists bool
	errQueryRow := stat.QueryRow(roleID, method, route).Scan(&exists)
	if errQueryRow != nil {
		if errQueryRow == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.New("RetrieveUserRoleAllowanceIsExist.errQueryRow : " + errQueryRow.Error())
	}

	return exists, nil

	//return true, nil
}
