package main

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
)

type SqlLogger struct {
	/*
		can conform interface by extended interface in struct
		don't need to implement all method in interface
	*/
	logger.Interface // conform logger interface
}

func (l SqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, _ := fc() // return sql statement that generate by gorm
	fmt.Printf("%v\n", sql)
}

var db *gorm.DB

func main() {
	dsn := "root:P@ssw0rd@tcp(13.76.163.73:3306)/gorm?parseTime=true"
	dial := mysql.Open(dsn) // return gorm.Dialector
	// dialector : database driver

	var err error
	db, err = gorm.Open(dial, &gorm.Config{
		Logger: &SqlLogger{},
		DryRun: false, // dry run if true
	})
	if err != nil {
		panic(err)
	}

	// gorm prefer convension over configuration
	// db.Migrator().CreateTable(Test{}) // create table
	db.AutoMigrate(Gender{}, Customer{}, Test{})

	// can use db.Raw() to write native query
}

func GetGenders() {
	/*
		Find: normal select (not return err),
		First: limit 1 order by pk ,
		Take: limit 1,
		Last: limit 1 order by pk desc
		First, Take, Last if not found return ErrRecordNotFound
	*/
	genders := []Gender{}
	tx := db.Order("id").Find(&genders) // if want to order by desc replace id to id desc
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	fmt.Println(genders)
}

func GetGenderByName(name string) {
	genders := []Gender{}
	tx := db.Find(&genders, "name = ?", name)
	// or db.Where("name = ?", name)Find(&genders)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	fmt.Println(genders)
}

func GetGender(id uint) {
	gender := Gender{}
	tx := db.First(&gender, id) // where id
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}
	fmt.Println(gender)
}

func CreateGender(name string) {
	gender := Gender{Name: name}
	tx := db.Create(&gender) // insert gender
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	fmt.Println(gender)
}

func UpdateGender(id uint, name string) {
	gender := Gender{}
	tx := db.First(&gender, id)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	gender.Name = name
	tx = db.Save(&gender)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}
}

func UpdateGender2(id uint, name string) {
	gender := Gender{ // set update field and new value (not zero value)
		Name: name,
	}
	// tx := db.Model(&Gender{}).Where("id = ?", id).Updates(gender)
	// update multi field (based on filed in struct)
	// Update() is other way to update one field
	tx := db.Model(&Gender{}).Where("id = @myid", sql.Named("myid", id)).Updates(gender) // use named argument
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	GetGender(id)
}

func DeleteGender(id uint) {
	// use gorm.Model it will softly delete
	// when select gorm will where deleted_at is null
	tx := db.Delete(&Gender{}, id)
	// use Unscoped() to delete permanently
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	fmt.Println("Deleted")
	GetGender(id)
}

func CreateCustomer(name string, genderID uint) {
	customer := Customer{Name: name, GenderID: genderID}
	tx := db.Create(&customer)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	fmt.Println(customer)
}

func GetCustomers() {
	customers := []Customer{}
	// tx := db.Preload("Gender").Find(&customers)
	// preload gender will query gender and fill in customer if you don't preload it will fill in id
	tx := db.Preload(clause.Associations).Find(&customers) // preload all association
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}

	for _, customer := range customers {
		fmt.Printf("%v|%v|%v", customer.ID, customer.Name, customer.Gender.Name)
	}
}

type Gender struct {
	ID   uint
	Name string `gorm:"unique;size(10)"`
}

type Customer struct {
	ID   uint
	Name string
	// gorm will create relation between customer and gender (customers.gender_id = genders.id)
	Gender   Gender
	GenderID uint
}

type Test struct {
	// table name was pluralize ex. gender to genders
	// table name and column name was converted to snake case
	ID   uint   // ID as primary key and auto increment in gorm
	Code uint   `gorm:"primaryKey;comment: code"`
	Name string `gorm:"column:firstname;type:varchar(50);unique;default:unknown;not null"`
	// varchar(50) can use size: 50
}

// set new table name by implement TableName() (Tabler interface)
func (t Test) TableName() string {
	return "MyTest"
}
