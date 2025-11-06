package main

import (
	"database/sql"
	"errors"
)

var ErrNoSuchParcel = errors.New("no such parcel")

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет строку в таблицу parcel
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))

	if err != nil {
		return 0, err
	}

	// возвращаем идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get реализует чтение строки из таблицы parcel по заданному number
// Возвращает эту строку
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", number))
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, ErrNoSuchParcel
		}
		return Parcel{}, err
	}

	return p, nil
}

// GetByClient реализует чтение строк из таблицы parcel по заданному client
// Возвращает эти строки
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel

	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Parcel{}

		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// SetStatus реализует обновление статуса в таблице parcel
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		return err
	}

	return nil
}

// SetAddres реализует обновление адреса в таблице parcel
// Менять адрес можно только если значение статуса registered
func (s ParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number and status = :status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}

	return nil
}

// Delete реализует удаление строки из таблицы parcel
// Удалять строку можно только если значение статуса registered
func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec("DELETE FROM parcel WHERE number = :number and status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoSuchParcel
	}

	return nil
}
