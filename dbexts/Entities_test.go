package dbexts

import (
	assertions "github.com/rafaelbfs/GoConvenience/Assertions"
	"testing"
)

func TestMakeUpdateStatement(t *testing.T) {
	someone := mkTestPerson()
	someoneNew := mkTestPerson()
	someoneNew.FirstName = "Jane"

	updates := MakeUpdateStatements(someone, someoneNew)

	assertions.AssertThat(t, len(updates)).EqualsTo(1)
	assertions.AssertThat(t, updates[0].Key).EqualsTo("first_name")
	assertions.AssertThat(t, updates[0].Value.(string)).EqualsTo("Jane")
}
