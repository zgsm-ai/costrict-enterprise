package env

import (
	"log"
	"testing"
)

func Test_initLogger(t *testing.T) {
	type args struct {
		debug   string
		logfile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Off", args{"Off", ""}, false},
		{"Dbg", args{"Dbg", ""}, false},
		{"Err", args{"Err", ""}, false},
		{"stderr_Err", args{"Err", ""}, false},
		{"file_Err", args{"Err", "test.log"}, false},
		{"multi_Err", args{"Err", "+test.log"}, false},
		{"stderr_Dbg", args{"Dbg", ""}, false},
		{"file_Dbg", args{"Dbg", "test.log"}, false},
		{"multi_Dbg", args{"Dbg", "+test.log"}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitLogger(tt.args.debug, tt.args.logfile); (err != nil) != tt.wantErr {
				t.Errorf("initLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
			log.Printf("%s: initLogger(debug=%s,logfile=%s)\n", tt.name, tt.args.debug, tt.args.logfile)
			LogDbg.Printf("%s: this is debug text\n", tt.name)
		})
	}
}
