package _115

import (
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/px-org/PanIndex/module"
)

func GetClient(account *module.Account) *driver.Pan115Client {
	return Sessions[account.Id]
}
