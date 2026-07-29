package main

import (
	"errors"
	"testing"
)

var _ error = QuotaError{}

func TestUploadFits(t *testing.T) {
	if err := Upload(50, 100, 30); err != nil {
		t.Errorf("Upload(50, 100, 30) = %v, ожидался nil — 80 из 100 влезает", err)
	}
}

func TestUploadExactLimit(t *testing.T) {
	if err := Upload(50, 100, 50); err != nil {
		t.Errorf("Upload(50, 100, 50) = %v, ожидался nil — ровно в лимит можно", err)
	}
}

func TestUploadOverflow(t *testing.T) {
	err := Upload(90, 100, 25)
	if err == nil {
		t.Fatalf("Upload(90, 100, 25) = nil, ожидалась ошибка — 115 из 100 не влезает")
	}
	if err.Error() != "квота исчерпана: 115 из 100" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "квота исчерпана: 115 из 100")
	}
}

func TestUploadFieldsViaAs(t *testing.T) {
	err := Upload(90, 100, 25)
	var qe QuotaError
	if !errors.As(err, &qe) {
		t.Fatalf("errors.As не нашла QuotaError — возвращайте именно этот тип")
	}
	if qe.Used != 115 || qe.Limit != 100 {
		t.Errorf("поля ошибки = {Used: %d, Limit: %d}, ожидалось {Used: 115, Limit: 100} — Used считается ПОСЛЕ загрузки", qe.Used, qe.Limit)
	}
}

func TestUploadWrappedStillFound(t *testing.T) {
	err := Upload(10, 10, 1)
	if err == nil {
		t.Fatalf("Upload(10, 10, 1) = nil, ожидалась ошибка — 11 из 10")
	}
	var qe QuotaError
	if !errors.As(err, &qe) {
		t.Fatalf("errors.As не нашла QuotaError")
	}
	if qe.Used-qe.Limit != 1 {
		t.Errorf("перебор = %d, ожидался 1", qe.Used-qe.Limit)
	}
}
