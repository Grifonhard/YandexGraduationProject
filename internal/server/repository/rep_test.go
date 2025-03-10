package repository

import (
	"context"
	"testing"

	"github.com/Grifonhard/YandexGraduationProject/DB/server"
	_ "github.com/jackc/pgx/v5/pgxpool"
)

var testDBConnString = server.CONN

func TestIntegration(t *testing.T) {
	if testDBConnString == "" {
		t.Skip("Пропускаем тест, т.к. не указана строка подключения TEST_DATABASE_URL")
	}

	// Создаём соединение c базой. Внутри вызова New() также вызывается CreateTables().
	db, err := New(testDBConnString)
	if err != nil {
		t.Fatalf("не удалось инициализировать репозиторий: %v", err)
	}
	defer db.Close()

	// Чтобы каждый тест не мешал друг другу, можно работать в рамках одного контекста
	ctx := context.Background()

	t.Run("User CRUD", func(t *testing.T) {
		// 1. Создаём пользователя
		userID, err := db.CreateUser(ctx, "test_user", "test_hash")
		if err != nil {
			t.Fatalf("CreateUser возвращает ошибку: %v", err)
		}

		// 2. Создаём пользователя с таким же именем (должна вернуться ошибка ErrDuplicate)
		_, err = db.CreateUser(ctx, "test_user", "another_hash")
		if err == nil {
			t.Fatalf("ожидалась ошибка при создании пользователя с тем же именем, но ошибки нет")
		}
		if err != nil && err.Error() != "user with such username already exists" && err.Error() != "user with such username repository: already exists" {
			// Уточняйте, как именно формируется ваша ошибка
			t.Logf("Ожидаемая ошибка: user with such username already exists, получили: %v", err)
		}

		// 3. Получаем пользователя
		gotUser, err := db.GetUser(ctx, userID)
		if err != nil {
			t.Fatalf("GetUser возвращает ошибку: %v", err)
		}
		if gotUser.Username != "test_user" {
			t.Errorf("у пользователя ожидался username = test_user, а вернулся: %s", gotUser.Username)
		}
		if gotUser.PasswordHash != "test_hash" {
			t.Errorf("ожидался password_hash = test_hash, а вернулся: %s", gotUser.PasswordHash)
		}

		// 4. Update пользователя
		err = db.UpdateUser(ctx, userID, "updated_user", "updated_hash")
		if err != nil {
			t.Fatalf("UpdateUser возвращает ошибку: %v", err)
		}

		// Снова получим и проверим
		gotUser, err = db.GetUser(ctx, userID)
		if err != nil {
			t.Fatalf("GetUser возвращает ошибку при повторном вызове: %v", err)
		}
		if gotUser.Username != "updated_user" || gotUser.PasswordHash != "updated_hash" {
			t.Errorf("данные после UpdateUser не совпадают с ожидаемыми")
		}

		// 5. ListUsers
		users, err := db.ListUsers(ctx)
		if err != nil {
			t.Fatalf("ListUsers возвращает ошибку: %v", err)
		}
		if len(users) == 0 {
			t.Error("ListUsers вернул пустой список, хотя мы создали пользователя")
		}

		// 6. DeleteUser
		err = db.DeleteUser(ctx, userID)
		if err != nil {
			t.Fatalf("DeleteUser возвращает ошибку: %v", err)
		}
		// Проверим, что пользователя реально больше нет
		_, err = db.GetUser(ctx, userID)
		if err == nil {
			t.Errorf("после удаления пользователь всё ещё существует, ожидали ошибку")
		}
	})

	t.Run("Services CRUD", func(t *testing.T) {
		// Сначала создадим нового пользователя, так как предыдущий был удалён
		userID, err := db.CreateUser(ctx, "user_for_service", "hash")
		if err != nil {
			t.Fatalf("CreateUser для сервисов вернул ошибку: %v", err)
		}

		// Создаём сервис
		serviceID, err := db.CreateService(ctx, userID, "test_service")
		if err != nil {
			t.Fatalf("CreateService вернул ошибку: %v", err)
		}

		// Пытаемся создать сервис с тем же именем
		_, err = db.CreateService(ctx, userID, "test_service")
		if err == nil {
			t.Fatalf("ожидали ошибку при повторном создании сервиса с одинаковым именем, но её нет")
		}

		// Проверим GetService
		gotService, err := db.GetService(ctx, serviceID)
		if err != nil {
			t.Fatalf("GetService вернул ошибку: %v", err)
		}
		if gotService.ServiceName != "test_service" {
			t.Errorf("ожидался service_name = test_service, получили: %s", gotService.ServiceName)
		}
		if gotService.UserID != userID {
			t.Errorf("ожидался userID = %d, получили: %d", userID, gotService.UserID)
		}

		// Проверим ListServices
		services, err := db.ListServices(ctx, userID)
		if err != nil {
			t.Fatalf("ListServices вернул ошибку: %v", err)
		}
		if len(services) != 1 {
			t.Errorf("ListServices должен вернуть ровно 1 сервис, вернул: %d", len(services))
		}

		// Обновим сервис
		err = db.UpdateService(ctx, serviceID, "new_service_name")
		if err != nil {
			t.Fatalf("UpdateService вернул ошибку: %v", err)
		}

		// Снова получим и проверим
		gotService, err = db.GetService(ctx, serviceID)
		if err != nil {
			t.Fatalf("GetService (после Update) вернул ошибку: %v", err)
		}
		if gotService.ServiceName != "new_service_name" {
			t.Errorf("после UpdateService имя не совпадает, получили: %s", gotService.ServiceName)
		}

		// Удалим сервис
		err = db.DeleteService(ctx, serviceID)
		if err != nil {
			t.Fatalf("DeleteService вернул ошибку: %v", err)
		}
		// Проверим, что сервис реально удалён
		_, err = db.GetService(ctx, serviceID)
		if err == nil {
			t.Error("после удаления сервис всё ещё существует, ожидали ошибку")
		}
	})

	t.Run("ServiceCreds CRUD", func(t *testing.T) {
		// Создадим пользователя и сервис
		userID, err := db.CreateUser(ctx, "user_for_creds", "hash")
		if err != nil {
			t.Fatalf("CreateUser вернул ошибку: %v", err)
		}
		serviceID, err := db.CreateService(ctx, userID, "creds_service")
		if err != nil {
			t.Fatalf("CreateService вернул ошибку: %v", err)
		}

		// Создаём учетные данные
		scID, err := db.CreateServiceCred(ctx, userID, serviceID, "login", []byte("encrypted"), []byte(`{"key":"value"}`))
		if err != nil {
			t.Fatalf("CreateServiceCred вернул ошибку: %v", err)
		}

		// Получаем
		sc, err := db.GetServiceCred(ctx, scID)
		if err != nil {
			t.Fatalf("GetServiceCred вернул ошибку: %v", err)
		}
		if sc.Login != "login" {
			t.Errorf("ожидался login = 'login', получили: %s", sc.Login)
		}
		// List
		allCreds, err := db.ListServiceCreds(ctx, userID, serviceID)
		if err != nil {
			t.Fatalf("ListServiceCreds вернул ошибку: %v", err)
		}
		if len(allCreds) != 1 {
			t.Errorf("ожидался 1 объект, получили: %d", len(allCreds))
		}

		// Обновим
		err = db.UpdateServiceCred(ctx, scID, "new_login", []byte("new_encrypted"), []byte(`{"new":"meta"}`))
		if err != nil {
			t.Fatalf("UpdateServiceCred вернул ошибку: %v", err)
		}

		sc, err = db.GetServiceCred(ctx, scID)
		if err != nil {
			t.Fatalf("GetServiceCred (после обновления) вернул ошибку: %v", err)
		}
		if sc.Login != "new_login" {
			t.Errorf("после обновления login != 'new_login', а '%s'", sc.Login)
		}

		// Удалим
		err = db.DeleteServiceCred(ctx, scID)
		if err != nil {
			t.Fatalf("DeleteServiceCred вернул ошибку: %v", err)
		}
		_, err = db.GetServiceCred(ctx, scID)
		if err == nil {
			t.Error("после удаления ServiceCred всё ещё существует")
		}
	})

	t.Run("TextData CRUD", func(t *testing.T) {
		userID, _ := db.CreateUser(ctx, "user_for_textdata", "hash")
		serviceID, _ := db.CreateService(ctx, userID, "textdata_service")

		tdID, err := db.CreateTextData(ctx, userID, serviceID, "some text", []byte(`{"test":"data"}`))
		if err != nil {
			t.Fatalf("CreateTextData: %v", err)
		}

		td, err := db.GetTextData(ctx, tdID)
		if err != nil {
			t.Fatalf("GetTextData: %v", err)
		}
		if td.TextData != "some text" {
			t.Errorf("ожидался TextData = 'some text', получили '%s'", td.TextData)
		}

		err = db.UpdateTextData(ctx, tdID, "new text", []byte(`{"changed":true}`))
		if err != nil {
			t.Fatalf("UpdateTextData: %v", err)
		}

		td, err = db.GetTextData(ctx, tdID)
		if err != nil {
			t.Fatalf("GetTextData (после апдейта): %v", err)
		}
		if td.TextData != "new text" {
			t.Errorf("после обновления TextData != 'new text'")
		}

		err = db.DeleteTextData(ctx, tdID)
		if err != nil {
			t.Fatalf("DeleteTextData: %v", err)
		}
		_, err = db.GetTextData(ctx, tdID)
		if err == nil {
			t.Error("после удаления TextData всё ещё существует")
		}
	})

	t.Run("TextBytes CRUD", func(t *testing.T) {
		userID, _ := db.CreateUser(ctx, "user_for_textbytes", "hash")
		serviceID, _ := db.CreateService(ctx, userID, "textbytes_service")

		tbID, err := db.CreateTextBytes(ctx, userID, serviceID, []byte("some bytes"), []byte(`{"meta":"tb"}`))
		if err != nil {
			t.Fatalf("CreateTextBytes: %v", err)
		}

		tb, err := db.GetTextBytes(ctx, tbID)
		if err != nil {
			t.Fatalf("GetTextBytes: %v", err)
		}
		if string(tb.TextBytes) != "some bytes" {
			t.Errorf("ожидались байты 'some bytes', получили '%s'", string(tb.TextBytes))
		}

		err = db.UpdateTextBytes(ctx, tbID, []byte("new bytes"), []byte(`{"meta":"changed"}`))
		if err != nil {
			t.Fatalf("UpdateTextBytes: %v", err)
		}

		tb, _ = db.GetTextBytes(ctx, tbID)
		if string(tb.TextBytes) != "new bytes" {
			t.Errorf("после обновления TextBytes != 'new bytes'")
		}

		err = db.DeleteTextBytes(ctx, tbID)
		if err != nil {
			t.Fatalf("DeleteTextBytes: %v", err)
		}
		_, err = db.GetTextBytes(ctx, tbID)
		if err == nil {
			t.Error("после удаления TextBytes всё ещё существует")
		}
	})

	t.Run("Cards CRUD", func(t *testing.T) {
		userID, _ := db.CreateUser(ctx, "user_for_cards", "hash")
		serviceID, _ := db.CreateService(ctx, userID, "cards_service")

		cardID, err := db.CreateCard(
			ctx,
			userID,
			serviceID,
			[]byte("enc_card_data"),
			"1234",
			12,
			2030,
			[]byte(`{"type":"visa"}`),
		)
		if err != nil {
			t.Fatalf("CreateCard: %v", err)
		}

		card, err := db.GetCard(ctx, cardID)
		if err != nil {
			t.Fatalf("GetCard: %v", err)
		}
		if card.CardLast != "1234" {
			t.Errorf("ожидался CardLast='1234', получили '%s'", card.CardLast)
		}

		err = db.UpdateCard(ctx, cardID, []byte("new_enc_data"), "9999", 1, 2025, []byte(`{"type":"mastercard"}`))
		if err != nil {
			t.Fatalf("UpdateCard: %v", err)
		}

		card, _ = db.GetCard(ctx, cardID)
		if card.CardLast != "9999" {
			t.Errorf("после обновления CardLast != '9999'")
		}

		err = db.DeleteCard(ctx, cardID)
		if err != nil {
			t.Fatalf("DeleteCard: %v", err)
		}
		_, err = db.GetCard(ctx, cardID)
		if err == nil {
			t.Error("после удаления Card всё ещё существует")
		}
	})
}
