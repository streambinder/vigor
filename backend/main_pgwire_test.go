package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// fakePGServer is a minimal PostgreSQL wire-protocol server for tests.
// It accepts any credentials, reports success for every statement, and
// answers the few query shapes the code under test reads results from:
// count queries return one row holding zero, pg_inherits listings
// return no rows, RETURNING clauses return one row of plausible values,
// and every other SELECT returns no rows. It is test code only.
type fakePGServer struct {
	listener net.Listener
	failOn   string
	wg       sync.WaitGroup
}

func startFakePGServer(t *testing.T) *fakePGServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake pg listen: %v", err)
	}
	server := &fakePGServer{listener: listener}
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			server.wg.Add(1)
			go func() {
				defer server.wg.Done()
				server.serve(conn)
			}()
		}
	}()
	t.Cleanup(func() {
		// The database pools under test keep their connections open, so
		// the serve goroutines stay blocked on reads; do not wait for
		// them, the test process reaps everything.
		_ = listener.Close()
	})
	return server
}

func (s *fakePGServer) dsn(dbname string) string {
	return fmt.Sprintf("postgres://user:pass@%s/%s?sslmode=disable", s.listener.Addr().String(), dbname)
}

type pgStatement struct {
	query  string
	params int
}

func (s *fakePGServer) serve(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	// Startup phase: answer SSL and GSS requests with 'N', then read the
	// startup packet itself.
	for {
		header := make([]byte, 8)
		if _, err := io.ReadFull(reader, header); err != nil {
			return
		}
		length := int(binary.BigEndian.Uint32(header[0:4])) - 8
		code := binary.BigEndian.Uint32(header[4:8])
		if code == 80877103 || code == 80877104 {
			if _, err := conn.Write([]byte{'N'}); err != nil {
				return
			}
			continue
		}
		if length > 0 {
			if _, err := io.CopyN(io.Discard, reader, int64(length)); err != nil {
				return
			}
		}
		break
	}
	statements := map[string]pgStatement{}
	portals := map[string]pgStatement{}
	skipUntilSync := false
	write := func(msgType byte, payload []byte) bool {
		frame := make([]byte, 5, len(payload)+5)
		frame[0] = msgType
		binary.BigEndian.PutUint32(frame[1:5], uint32(len(payload)+4))
		frame = append(frame, payload...)
		_, err := conn.Write(frame)
		return err == nil
	}
	ready := func() bool { return write('Z', []byte{'I'}) }
	cstr := func(v string) []byte { return append([]byte(v), 0) }
	// AuthenticationOk, the parameter statuses pgx reads, BackendKeyData,
	// then ReadyForQuery.
	if !write('R', []byte{0, 0, 0, 0}) {
		return
	}
	for _, kv := range [][2]string{
		{"server_version", "16.0"},
		{"server_encoding", "UTF8"},
		{"client_encoding", "UTF8"},
		{"DateStyle", "ISO, MDY"},
		{"integer_datetimes", "on"},
		{"standard_conforming_strings", "on"},
		{"TimeZone", "UTC"},
	} {
		if !write('S', append(cstr(kv[0]), cstr(kv[1])...)) {
			return
		}
	}
	keyData := make([]byte, 8)
	binary.BigEndian.PutUint32(keyData[0:4], 42)
	binary.BigEndian.PutUint32(keyData[4:8], 42)
	if !write('K', keyData) || !ready() {
		return
	}

	fieldsFor := func(query string) []string {
		lower := strings.ToLower(query)
		if strings.Contains(lower, "returning") {
			idx := strings.LastIndex(lower, "returning")
			rest := strings.TrimSpace(query[idx+len("returning"):])
			var fields []string
			for _, part := range strings.Split(rest, ",") {
				name := strings.Trim(strings.TrimSpace(part), `"`)
				if name != "" && !strings.ContainsAny(name, " ()") {
					fields = append(fields, name)
				}
			}
			if len(fields) > 0 {
				return fields
			}
		}
		if strings.Contains(lower, "count(") || strings.Contains(lower, "count (*)") {
			return []string{"count"}
		}
		if strings.Contains(lower, "pg_inherits") {
			return []string{"inhrelid"}
		}
		if strings.Contains(lower, "version()") {
			return []string{"version"}
		}
		return []string{"value"}
	}
	rowDescription := func(fields []string) []byte {
		payload := make([]byte, 2)
		binary.BigEndian.PutUint16(payload, uint16(len(fields)))
		for _, field := range fields {
			payload = append(payload, cstr(field)...)
			meta := make([]byte, 18)
			oid := uint32(25)
			if field == "count" {
				oid = 20
			}
			binary.BigEndian.PutUint32(meta[6:10], oid)
			binary.BigEndian.PutUint16(meta[10:12], 0xffff)
			binary.BigEndian.PutUint32(meta[12:16], 0xffffffff)
			payload = append(payload, meta...)
		}
		return payload
	}
	dataRow := func(values []string) []byte {
		payload := make([]byte, 2)
		binary.BigEndian.PutUint16(payload, uint16(len(values)))
		for _, v := range values {
			lenBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBytes, uint32(len(v)))
			payload = append(payload, lenBytes...)
			payload = append(payload, v...)
		}
		return payload
	}
	defaultValue := func(field string) string {
		switch {
		case field == "count":
			return "0"
		case strings.HasSuffix(field, "_at") || field == "time":
			return "2026-01-01T00:00:00Z"
		case field == "id" || strings.HasSuffix(field, "_id"):
			return "00000000-0000-0000-0000-000000000001"
		default:
			return ""
		}
	}
	rowsFor := func(query string) [][]string {
		lower := strings.ToLower(query)
		fields := fieldsFor(query)
		if strings.Contains(lower, "count(") {
			return [][]string{{"0"}}
		}
		if strings.Contains(lower, "returning") {
			row := make([]string, len(fields))
			for i, f := range fields {
				row[i] = defaultValue(f)
			}
			return [][]string{row}
		}
		if strings.Contains(lower, "version()") {
			return [][]string{{"PostgreSQL 16.0"}}
		}
		return nil
	}
	commandTag := func(query string, rows int) string {
		verb := strings.ToUpper(strings.TrimSpace(query))
		switch {
		case strings.HasPrefix(verb, "SELECT"):
			return "SELECT " + strconv.Itoa(rows)
		case strings.HasPrefix(verb, "INSERT"):
			return "INSERT 0 " + strconv.Itoa(rows)
		case strings.HasPrefix(verb, "UPDATE"):
			return "UPDATE " + strconv.Itoa(rows)
		case strings.HasPrefix(verb, "DELETE"):
			return "DELETE " + strconv.Itoa(rows)
		case strings.HasPrefix(verb, "CREATE TABLE"):
			return "CREATE TABLE"
		case strings.HasPrefix(verb, "CREATE INDEX"):
			return "CREATE INDEX"
		case strings.HasPrefix(verb, "CREATE SCHEMA"):
			return "CREATE SCHEMA"
		case strings.HasPrefix(verb, "CREATE EXTENSION"):
			return "CREATE EXTENSION"
		case strings.HasPrefix(verb, "DROP TABLE"):
			return "DROP TABLE"
		default:
			return "OK"
		}
	}
	fail := func(query string) bool {
		return s.failOn != "" && strings.Contains(query, s.failOn)
	}
	errorResponse := func() bool {
		payload := append([]byte{'S'}, cstr("ERROR")...)
		payload = append(payload, 'V')
		payload = append(payload, cstr("ERROR")...)
		payload = append(payload, 'C')
		payload = append(payload, cstr("42601")...)
		payload = append(payload, 'M')
		payload = append(payload, cstr("fake server failure")...)
		payload = append(payload, 0)
		return write('E', payload)
	}
	execute := func(query string) bool {
		rows := rowsFor(query)
		for _, row := range rows {
			if !write('D', dataRow(row)) {
				return false
			}
		}
		return write('C', cstr(commandTag(query, len(rows))))
	}

	for {
		msgType, err := reader.ReadByte()
		if err != nil {
			return
		}
		lenHeader := make([]byte, 4)
		if _, err := io.ReadFull(reader, lenHeader); err != nil {
			return
		}
		length := int(binary.BigEndian.Uint32(lenHeader)) - 4
		payload := make([]byte, length)
		if length > 0 {
			if _, err := io.ReadFull(reader, payload); err != nil {
				return
			}
		}
		if skipUntilSync && msgType != 'S' {
			continue
		}
		switch msgType {
		case 'X':
			return
		case 'Q':
			query := strings.TrimSuffix(string(payload), "\x00")
			if strings.TrimSpace(query) == "" {
				if !write('I', nil) || !ready() {
					return
				}
				continue
			}
			if fail(query) {
				if !errorResponse() || !ready() {
					return
				}
				continue
			}
			fields := fieldsFor(query)
			rows := rowsFor(query)
			if len(rows) > 0 || strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT") {
				if !write('T', rowDescription(fields)) {
					return
				}
			}
			if !execute(query) || !ready() {
				return
			}
		case 'P':
			rest := payload
			name, rest := readCString(rest)
			query, _ := readCString(rest)
			// pgx declares zero parameter types at Parse time and lets
			// the server infer them, so count the $n placeholders in
			// the statement text instead.
			count := 0
			for _, match := range placeholders(query) {
				if match > count {
					count = match
				}
			}
			statements[name] = pgStatement{query: query, params: count}
			if fail(query) {
				if !errorResponse() {
					return
				}
				skipUntilSync = true
				continue
			}
			if !write('1', nil) {
				return
			}
		case 'D':
			if len(payload) < 1 {
				return
			}
			kind, name := payload[0], strings.TrimSuffix(string(payload[1:]), "\x00")
			var stmt pgStatement
			if kind == 'S' {
				stmt = statements[name]
			} else {
				stmt = portals[name]
			}
			// ParameterDescription echoing the Parse-time parameter count.
			params := make([]byte, 2)
			binary.BigEndian.PutUint16(params, uint16(stmt.params))
			for i := 0; i < stmt.params; i++ {
				oid := make([]byte, 4)
				binary.BigEndian.PutUint32(oid, 0)
				params = append(params, oid...)
			}
			if !write('t', params) {
				return
			}
			if !write('T', rowDescription(fieldsFor(stmt.query))) {
				return
			}
		case 'B':
			rest := payload
			portal, rest := readCString(rest)
			stmtName, _ := readCString(rest)
			portals[portal] = statements[stmtName]
			if !write('2', nil) {
				return
			}
		case 'E':
			portal, _ := readCString(payload)
			stmt := portals[portal]
			if !execute(stmt.query) {
				return
			}
		case 'C':
			if !write('3', nil) {
				return
			}
		case 'H':
			// flush only: nothing is buffered
		case 'S':
			skipUntilSync = false
			if !ready() {
				return
			}
		default:
			// CopyData, FunctionCall and friends are not used by the code
			// under test; ignore them.
		}
	}
}

// placeholders returns the indices of the $n placeholders in a query.
func placeholders(query string) []int {
	var out []int
	for i := 0; i < len(query); i++ {
		if query[i] != '$' {
			continue
		}
		j := i + 1
		n := 0
		for j < len(query) && query[j] >= '0' && query[j] <= '9' {
			n = n*10 + int(query[j]-'0')
			j++
		}
		if j > i+1 {
			out = append(out, n)
			i = j - 1
		}
	}
	return out
}

func readCString(data []byte) (string, []byte) {
	idx := strings.IndexByte(string(data), 0)
	if idx < 0 {
		return string(data), nil
	}
	return string(data[:idx]), data[idx+1:]
}
