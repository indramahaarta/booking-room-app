package repository

import (
	"time"

	"github.com/indramahaarta/bookings/internal/model"
)

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res model.Reservation) (int, error)
	InsertRoomRestriction(r model.RoomRestriction) error
	SearchAvailbilityByDatesByRoomId(roomId int, startDate, endDate time.Time) (bool, error)
	SearchRoomAvailbilityByDate(startDate, endDate time.Time) ([]model.Room, error)
	GetRoomByID(id int) (model.Room, error)
	GetUserById(id int) (model.User, error)
	UpdateUser(u model.User) error
	Authenticate(email, password string) (int, string, error)
}
