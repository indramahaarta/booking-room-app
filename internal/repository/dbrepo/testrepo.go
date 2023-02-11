package dbrepo

import (
	"time"

	"github.com/indramahaarta/bookings/internal/model"
)

func (m *testPostgresDBRepo) AllUsers() bool {
	return true
}

func (m *testPostgresDBRepo) InsertReservation(res model.Reservation) (int, error) {
	return 1, nil
}

func (m *testPostgresDBRepo) InsertRoomRestriction(r model.RoomRestriction) error {
	return nil
}

func (m *testPostgresDBRepo) SearchAvailbilityByDatesByRoomId(roomId int, startDate, endDate time.Time) (bool, error) {
	return true, nil

}

func (m *testPostgresDBRepo) SearchRoomAvailbilityByDate(startDate, endDate time.Time) ([]model.Room, error) {
	var rooms []model.Room
	return rooms, nil

}

func (m *testPostgresDBRepo) GetRoomByID(id int) (model.Room, error) {
	var room model.Room
	return room, nil
}

func (m *testPostgresDBRepo) UpdateUser(u model.User) error {
	return nil

}

func (m *testPostgresDBRepo) Authenticate(email, password string) (int, string, error) {
	return 0, "", nil

}

func (m *testPostgresDBRepo) GetUserById(id int) (model.User, error) {
	return model.User{}, nil
}
