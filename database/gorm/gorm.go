package gorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/fatih/color"
	"github.com/maxpowel/dislet"
	"github.com/maxpowel/wiphonego"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"

)

type Config struct {
	Dialect string `default:"valor"`
	Uri string `default:"PUERTOOO"`
}


func NewConnection(dialect string, uri string) *gorm.DB {
	db, err := gorm.Open(dialect, uri)
	//fmt.Println(uri)
	//db.LogMode(true)
	fmt.Println("CREANDO CONEXION")
	if err != nil {
		log.Fatal(err)
	}

	//db.Model(&wiphonego.UserDevice{}).Related(&wiphonego.UserDeviceConsumption{})
	db.Model(&wiphonego.Device{}).Related(&wiphonego.PhoneLine{})
	db.Model(&wiphonego.PhoneLineConsumption{}).Related(&wiphonego.PhoneLine{})
	db.Model(&wiphonego.PhoneLine{}).Related(&wiphonego.Device{})
	db.Model(&wiphonego.PhoneLine{}).Related(&wiphonego.Credentials{})
	db.Model(&wiphonego.Credentials{}).Related(&wiphonego.Operator{})

	db.AutoMigrate(&wiphonego.Device{}, &wiphonego.PhoneLine{}, &wiphonego.PhoneLineConsumption{}, &wiphonego.Credentials{}, &wiphonego.Operator{})

	//db.AutoMigrate(&User{})
	/*u := NewUser()
	PlainPassword(&u, "123456")
	db.Create(&u)
	db.Commit()*/
	return db
}


func Bootstrap(k *dislet.Kernel) {
	//fmt.Println("DATABASE BOOT")
	mapping := k.Config.Mapping
	mapping["database"] = &Config{}

	var baz dislet.OnKernelReady = func(k *dislet.Kernel){
		color.Green("Booting database")
		conf := k.Config.Mapping["database"].(*Config)


		k.Container.RegisterType("database", NewConnection, conf.Dialect, conf.Uri)
	}
	k.Subscribe(baz)

}
