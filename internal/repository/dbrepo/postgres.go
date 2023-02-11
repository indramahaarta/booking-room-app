package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/indramahaarta/bookings/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res model.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var id int

	stmt := `
		insert into reservations
		(first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;
	`

	err := m.DB.QueryRowContext(ctx, stmt, res.FirstName, res.LastName, res.Email, res.Phone, res.StartDate, res.EndDate, res.RoomId, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}

	log.Println("Success: Reservation has been inserted to database!")

	return id, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(r model.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stmt := `
		insert into room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err := m.DB.ExecContext(ctx, stmt, r.StartDate, r.EndDate, r.RoomId, r.ReservationId, r.RestrictionId, time.Now(), time.Now())
	if err != nil {
		return err
	}

	log.Println("Success: Room Restriction has been made!")

	return nil
}

func (m *postgresDBRepo) SearchAvailbilityByDatesByRoomId(roomId int, startDate, endDate time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stmt := `
		select count(id)
		from room_restrictions
		where $1 = room_id and $2 < end_date and $3 > start_date;
	`

	row := m.DB.QueryRowContext(ctx, stmt, roomId, startDate, endDate)

	var numRows int
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil

}

func (m *postgresDBRepo) SearchRoomAvailbilityByDate(startDate, endDate time.Time) ([]model.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var rooms []model.Room

	stmt := `
		select r.id, r.name
		from rooms r
		where r.id not in (
			select rr.room_id
			from room_restrictions rr
			where $1 < rr.end_date and $2 > rr.start_date
		)
	`

	rows, err := m.DB.QueryContext(ctx, stmt, startDate, endDate)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room model.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil

}

func (m *postgresDBRepo) GetRoomByID(id int) (model.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var room model.Room
	stmt := `
		select id, name, created_at, updated_at
		from rooms where id = $1
	`

	row := m.DB.QueryRowContext(ctx, stmt, id)
	row.Scan(&room.ID, &room.RoomName, &room.CreateAt, &room.UpdateAt)
	if err := row.Err(); err != nil {
		return room, err
	}

	return room, nil
}

func (m *postgresDBRepo) GetUserById(id int) (model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select id, first_name, last_name, email, password, acces_level, created_at, updated_at 
	from users where id=$1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u model.User
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.AccessLevel, &u.CreateAt, &u.UpdateAt)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *postgresDBRepo) UpdateUser(u model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name=$1, last_name=$2, email=$3, access_level=$4, update_at=$5
		where id=$6
	`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now())
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (m *postgresDBRepo) Authenticate(email, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string
	query := `select id, password from users where email=$1`

	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}
