package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestEncodeFull(t *testing.T) {
	u := User{Name: "gopher", Email: "go@example.com", Age: 3,
		Friends: []string{"rob", "ken"}, Token: "секрет"}
	got, err := EncodeUser(u)
	if err != nil {
		t.Fatalf("EncodeUser(полный профиль) вернула ошибку: %v", err)
	}
	want := `{"name":"gopher","email":"go@example.com","age":3,"friends":["rob","ken"]}`
	if got != want {
		t.Errorf("EncodeUser(полный профиль) = %s, ожидалось %s (проверьте теги: имена строчными, порядок полей не менять)", got, want)
	}
}

func TestEncodeOmitsEmpty(t *testing.T) {
	got, err := EncodeUser(User{Name: "нео"})
	if err != nil {
		t.Fatalf("EncodeUser(только имя) вернула ошибку: %v", err)
	}
	want := `{"name":"нео"}`
	if got != want {
		t.Errorf("EncodeUser(только имя) = %s, ожидалось %s — пустые email/age/friends должен спрятать omitempty", got, want)
	}
}

func TestTokenNeverLeaks(t *testing.T) {
	got, err := EncodeUser(User{Name: "gopher", Token: "sekret-123"})
	if err != nil {
		t.Fatalf("EncodeUser вернула ошибку: %v", err)
	}
	if strings.Contains(got, "sekret-123") || strings.Contains(strings.ToLower(got), "token") {
		t.Errorf("токен утёк в JSON: %s — поле Token помечается тегом json:\"-\"", got)
	}
}

func TestDecodeFull(t *testing.T) {
	got, err := DecodeUser(`{"name":"gopher","email":"go@example.com","age":3,"friends":["rob","ken"]}`)
	if err != nil {
		t.Fatalf("DecodeUser(полный JSON) вернула ошибку: %v", err)
	}
	if got.Name != "gopher" || got.Email != "go@example.com" || got.Age != 3 {
		t.Errorf("DecodeUser: Name=%q Email=%q Age=%d, ожидалось gopher / go@example.com / 3", got.Name, got.Email, got.Age)
	}
	if !reflect.DeepEqual(got.Friends, []string{"rob", "ken"}) {
		t.Errorf("DecodeUser: Friends=%v, ожидалось [rob ken]", got.Friends)
	}
}

func TestDecodeMissingFields(t *testing.T) {
	got, err := DecodeUser(`{"name":"ann","age":19}`)
	if err != nil {
		t.Fatalf("DecodeUser вернула ошибку: %v", err)
	}
	if got.Name != "ann" || got.Age != 19 {
		t.Errorf("DecodeUser: Name=%q Age=%d, ожидалось ann / 19", got.Name, got.Age)
	}
	if got.Email != "" || len(got.Friends) != 0 {
		t.Errorf("отсутствующие в JSON поля должны остаться нулевыми: Email=%q Friends=%v", got.Email, got.Friends)
	}
}

func TestDecodeIgnoresToken(t *testing.T) {
	got, err := DecodeUser(`{"name":"ева","token":"взлом"}`)
	if err != nil {
		t.Fatalf("DecodeUser вернула ошибку: %v", err)
	}
	if got.Token != "" {
		t.Errorf("Token=%q после разбора, ожидалось пустое — тег json:\"-\" игнорирует поле в обе стороны", got.Token)
	}
}

func TestDecodeBadJSON(t *testing.T) {
	_, err := DecodeUser(`{"name": кривой json`)
	if err == nil {
		t.Errorf("DecodeUser(кривой JSON) вернула nil-ошибку — ошибку json.Unmarshal нужно пробросить наружу")
	}
}
