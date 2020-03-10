package datasource

import (
	"context"
	"fmt"

	"github.com/dictyBase/go-genproto/dictybaseapis/api/jsonapi"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
)

type User struct {
	client user.UserServiceClient
}

func (u *User) UsersInOrder(ord *order.Order) (map[string]*user.User, error) {
	m := make(map[string]*user.User)
	pu, err := u.client.GetUserByEmail(
		context.Background(),
		&jsonapi.GetEmailRequest{Email: ord.Data.Attributes.Payer},
	)
	if err != nil {
		return m, fmt.Errorf("error in retrieving payer %s", err)
	}
	su, err := u.client.GetUserByEmail(
		context.Background(),
		&jsonapi.GetEmailRequest{Email: ord.Data.Attributes.Consumer},
	)
	if err != nil {
		return m, fmt.Errorf("error in retrieving shipper %s", err)
	}
	m["payer"] = pu
	m["shipper"] = su
	return m, nil
}
