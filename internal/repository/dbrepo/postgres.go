package dbrepo

import (
	"context"
	"github.com/raul/BookingSystem/internal/models"
	"time"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newId int

	stmt := `insert into reservations 
    	(first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomId,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}

	return newId, err
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions 
    	(start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomId,
		r.ReservationId,
		r.RestrictionId,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return err
}

// SearchAvailabilityByDateByRoomId returns true if availability exists for roomId and false otherwise
func (m *postgresDBRepo) SearchAvailabilityByDateByRoomId(start, end time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select count(id)
	from room_restrictions
	where room_id = $1 and start_date >= CURRENT_DATE and not ((
	    -- the reservation is in between
	    $2  >= start_date and $3 <= end_date ) or (
	    -- the reservation encapsulates another reservation
	    $2 < start_date and $3 > end_date ) or (
	    -- the reservation intersects another reservation by start date
	    $3 > start_date and $3 < end_date ) or (
	    -- the reservation intersects another reservation by end date
	    $2 < end_date and $3 >= end_date ))`

	var numRows int

	row := m.DB.QueryRowContext(ctx, query,
		roomId,
		start,
		end,
	)

	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}
	if row.Err() != nil {
		return false, row.Err()
	}

	if numRows == 0 {
		return false, nil
	}

	return true, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms if any for given date range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m.App.InfoLog.Println("SEARCHING FOR AVAILABILITY: ", start, end)

	query := `
	select r.id, r.room_name
	from rooms as r
	where r.id not in (
	    select room_id
	    from room_restrictions
		where 
			-- the reservation is in between
			($1  >= start_date and $2 <= end_date ) or (
			-- the reservation encapsulates another reservation
			$1 < start_date and $2 > end_date ) or (
			-- the reservation intersects another reservation by start date
			$2 > start_date and $2 < end_date  ) or (
			-- the reservation intersects another reservation by end date
			$1 < end_date and $2 >= end_date)
	    )`

	var rooms []models.Room

	rows, err := m.DB.QueryContext(ctx, query,
		start,
		end)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.Id, &room.RoomName)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetRoomById returns the room with the specified id
func (m *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query,
		id,
	)

	err := row.Scan(&room.Id, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return models.Room{}, err
	}
	if row.Err() != nil {
		return models.Room{}, row.Err()
	}

	return room, nil
}
