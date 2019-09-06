package main

import (
	"reflect"
	"testing"
)

func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	if reflect.DeepEqual(a, b) {
		return
	}

	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func TestIsLAN(t *testing.T) {
	AssertEqual(t, isLAN([]string{"LAN"}), true)
	AssertEqual(t, isLAN([]string{"ALAN"}), false)
	AssertEqual(t, isLAN([]string{"WAN"}), false)
}

func TestIsActive(t *testing.T) {
	AssertEqual(t, isActive([]string{"", " Link Down"}), false)
	AssertEqual(t, isActive([]string{"", "link down"}), false)
	AssertEqual(t, isActive([]string{"", "100/10"}), true)
}

func TestFindAllSubmatchGroups(t *testing.T) {
	AssertEqual(t, findAllSubmatchGroups(`\d+ (\d+) \d+`, "123 456 789", 1), []string{"456"})
	AssertEqual(t, findAllSubmatchGroups(`(\d+)`, "123 456 789", 1), []string{"123", "456", "789"})
	AssertEqual(t, findAllSubmatchGroups(`(\d+) (\d+) (\d+)`, "123 456 789", 3), []string{"789"})
}
