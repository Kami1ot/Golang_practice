package main

import "testing"

func TestEngineDescribe(t *testing.T) {
	e := Engine{Power: 150}
	want := "двигатель 150 л.с."
	if got := e.Describe(); got != want {
		t.Errorf("Engine{Power: 150}.Describe() = %q, ожидалось %q", got, want)
	}
}

func TestEngineDescribeOtherPower(t *testing.T) {
	e := Engine{Power: 75}
	want := "двигатель 75 л.с."
	if got := e.Describe(); got != want {
		t.Errorf("Engine{Power: 75}.Describe() = %q, ожидалось %q — число должно браться из поля Power", got, want)
	}
}

func TestCarDescribePromoted(t *testing.T) {
	c := Car{Engine: Engine{Power: 200}, Brand: "Gopher"}
	want := "двигатель 200 л.с."
	if got := c.Describe(); got != want {
		t.Errorf("c.Describe() = %q, ожидалось %q — метод Describe должен продвигаться из Engine и вызываться прямо у Car", got, want)
	}
}

func TestCarFullName(t *testing.T) {
	c := Car{Engine: Engine{Power: 90}, Brand: "Gopher"}
	want := "Gopher, двигатель 90 л.с."
	if got := c.FullName(); got != want {
		t.Errorf("c.FullName() = %q, ожидалось %q — формат: марка, запятая с пробелом, описание двигателя", got, want)
	}
}

func TestPromotedFieldAccess(t *testing.T) {
	c := Car{Engine: Engine{Power: 120}, Brand: "GoMobile"}
	if c.Power != 120 {
		t.Errorf("c.Power = %d, ожидалось 120 — продвинутое поле должно читаться напрямую у Car", c.Power)
	}
	if c.Power != c.Engine.Power {
		t.Errorf("c.Power (%d) и c.Engine.Power (%d) должны быть одним и тем же полем — не меняйте объявления типов", c.Power, c.Engine.Power)
	}
}
