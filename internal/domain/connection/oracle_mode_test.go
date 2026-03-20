package connection

import "testing"

func TestOracleConnectionValidate_BasicServiceName(t *testing.T) {
	conn := &OracleConnection{
		BaseConnection: BaseConnection{Name: "Oracle Basic"},
		Host:           "db.local",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "system",
		ConnectType:    "basic",
		IdentifierType: "service_name",
	}

	if err := conn.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestOracleConnectionValidate_BasicSID(t *testing.T) {
	conn := &OracleConnection{
		BaseConnection: BaseConnection{Name: "Oracle SID"},
		Host:           "db.local",
		Port:           1521,
		SID:            "ORCLSID",
		Username:       "system",
		ConnectType:    "basic",
		IdentifierType: "sid",
	}

	if err := conn.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestOracleConnectionValidate_TNS(t *testing.T) {
	conn := &OracleConnection{
		BaseConnection: BaseConnection{Name: "Oracle TNS"},
		TNSName:        "ORCLCDB_HIGH",
		Username:       "system",
		ConnectType:    "tns",
	}

	if err := conn.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestOracleConnectionValidate_TNSDoesNotRequireBasicFields(t *testing.T) {
	conn := &OracleConnection{
		BaseConnection: BaseConnection{Name: "Oracle TNS"},
		TNSName:        "ORCLCDB_HIGH",
		Username:       "system",
		ConnectType:    "tns",
	}

	if err := conn.Validate(); err != nil {
		t.Fatalf("Validate() should allow TNS without host/port/service/sid, got %v", err)
	}
}

func TestOracleConnectionGetDSNUsesTNSName(t *testing.T) {
	conn := &OracleConnection{
		Username:    "system",
		Password:    "secret",
		TNSName:     "ORCLCDB_HIGH",
		ConnectType: "tns",
	}

	if got := conn.GetDSNWithPassword(); got != "oracle://system:secret@ORCLCDB_HIGH" {
		t.Fatalf("GetDSNWithPassword() = %q", got)
	}
}
