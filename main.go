package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// Статус посылки
const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

// Parcel структура посылки
type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

// ParcelService сервис для управления посылками
type ParcelService struct {
	store ParcelStore
}

// NewParcelService создает новый экземпляр ParcelService
func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

// Register создает новую посылку, заполняет поля структуры и добавляет запись в таблицу
// Возвращает созданную посылку или ошибку в случае неудачи
func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

// PrintClientParcels выводит информацию о всех посылках данного клиента
// Возвращает ошибку если не удалось получить данные из таблицы
func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", client)
	for _, parcel := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
	}
	fmt.Println()

	return nil
}

// NextStatus меняет статус посылки:
// зарегистрирована -> отправлена -> доставлена
// Для доставленной посылки статус не меняется
// Возвращает ошибку если посылка не найдена
func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", number, nextStatus)

	return s.store.SetStatus(number, nextStatus)
}

// ChangeAdress изменяет адрес доставки посылки
// Возвращает ошику если посылка не найдена или не удалось обновить адрес
func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

// Delete удаляет посылку по номеру
// Возвращает ошику если посылка не найдена или не удалось удалить запись
func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	// настраиваем подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db) // создаем объект ParcelStore функцией NewParcelStore
	service := NewParcelService(store)

	// регистрация посылки
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// изменение адреса
	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	// изменение статуса
	err = service.NextStatus(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// вывод посылок клиента
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// попытка удаления отправленной посылки
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// вывод посылок клиента
	// предыдущая посылка не должна удалиться, т.к. её статус НЕ «зарегистрирована»
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// регистрация новой посылки
	p, err = service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// удаление новой посылки
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// вывод посылок клиента
	// здесь не должно быть последней посылки, т.к. она должна была успешно удалиться
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}
}
