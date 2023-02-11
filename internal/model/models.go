package model

import (
	"time"
)

type User struct {
	ID          int
	FirstName    string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreateAt    time.Time
	UpdateAt    time.Time
}

type Room struct {
	ID       int
	RoomName string
	CreateAt time.Time
	UpdateAt time.Time
}

type Restriction struct {
	ID              int
	RestrictionName string
	CreateAt        time.Time
	UpdateAt        time.Time
}

type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomId    int
	CreateAt  time.Time
	UpdateAt  time.Time
	Room      Room
}

type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomId        int
	ReservationId int
	RestrictionId int
	CreateAt      time.Time
	UpdateAt      time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}
