package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/mikeocool/bbox/core"
)

// WKT geometry types for detection
var wktGeometryTypes = []string{
	"POINT", "LINESTRING", "POLYGON",
	"MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON",
	"GEOMETRYCOLLECTION",
}

// SniffWkt checks if the data looks like Well-Known Text (WKT) or Extended WKT (EWKT) format
func SniffWkt(data []byte) bool {
	if len(data) < 5 { // Minimum: "POINT" = 5 characters
		return false
	}

	// Convert to uppercase string for pattern matching
	dataStr := strings.ToUpper(strings.TrimSpace(string(data)))

	// Check for EWKT format with SRID prefix
	if strings.HasPrefix(dataStr, "SRID=") {
		// Find the semicolon that separates SRID from geometry
		semicolonPos := strings.Index(dataStr, ";")
		if semicolonPos == -1 {
			return false
		}
		// Extract the geometry part after the semicolon
		geometryPart := strings.TrimSpace(dataStr[semicolonPos+1:])
		dataStr = geometryPart
	}

	// Must start with a known geometry type
	for _, geomType := range wktGeometryTypes {
		if strings.HasPrefix(dataStr, geomType) {
			// Check if followed by whitespace or opening parenthesis
			if len(dataStr) > len(geomType) {
				nextChar := dataStr[len(geomType)]
				if nextChar == ' ' || nextChar == '\t' || nextChar == '\n' || nextChar == '(' {
					return true
				}
			}
		}
	}

	return false
}

// Token types for WKT parsing
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenWord
	tokenNumber
	tokenLParen
	tokenRParen
	tokenComma
	tokenSemicolon
	tokenEquals
	tokenError
)

type token struct {
	typ   tokenType
	value string
}

// Lexer for streaming WKT parsing
type wktLexer struct {
	reader *bufio.Reader
	err    error
}

func newWktLexer(r io.Reader) *wktLexer {
	return &wktLexer{
		reader: bufio.NewReader(r),
	}
}

func (l *wktLexer) nextToken() token {
	if l.err != nil {
		return token{typ: tokenError, value: l.err.Error()}
	}

	// Skip whitespace
	for {
		ch, err := l.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				return token{typ: tokenEOF}
			}
			l.err = err
			return token{typ: tokenError, value: err.Error()}
		}

		if !unicode.IsSpace(rune(ch)) {
			l.reader.UnreadByte()
			break
		}
	}

	ch, err := l.reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return token{typ: tokenEOF}
		}
		l.err = err
		return token{typ: tokenError, value: err.Error()}
	}

	switch ch {
	case '(':
		return token{typ: tokenLParen, value: "("}
	case ')':
		return token{typ: tokenRParen, value: ")"}
	case ',':
		return token{typ: tokenComma, value: ","}
	case ';':
		return token{typ: tokenSemicolon, value: ";"}
	case '=':
		return token{typ: tokenEquals, value: "="}
	default:
		l.reader.UnreadByte()
		if unicode.IsLetter(rune(ch)) {
			return l.readWord()
		} else if unicode.IsDigit(rune(ch)) || ch == '-' || ch == '+' || ch == '.' {
			return l.readNumber()
		} else {
			l.err = fmt.Errorf("unexpected character: %c", ch)
			return token{typ: tokenError, value: l.err.Error()}
		}
	}
}

func (l *wktLexer) readWord() token {
	var word strings.Builder
	for {
		ch, err := l.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			l.err = err
			return token{typ: tokenError, value: err.Error()}
		}

		if unicode.IsLetter(rune(ch)) {
			word.WriteByte(ch)
		} else {
			l.reader.UnreadByte()
			break
		}
	}
	return token{typ: tokenWord, value: strings.ToUpper(word.String())}
}

func (l *wktLexer) readNumber() token {
	var number strings.Builder
	for {
		ch, err := l.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			l.err = err
			return token{typ: tokenError, value: err.Error()}
		}

		if unicode.IsDigit(rune(ch)) || ch == '.' || ch == '-' || ch == '+' || ch == 'e' || ch == 'E' {
			number.WriteByte(ch)
		} else {
			l.reader.UnreadByte()
			break
		}
	}
	return token{typ: tokenNumber, value: number.String()}
}

// WKT Parser
type wktParser struct {
	lexer        *wktLexer
	currentToken token
	bbox         bboxAccumulator
	srid         int
}

type bboxAccumulator struct {
	minX, minY, maxX, maxY float64
	initialized            bool
}

func (b *bboxAccumulator) addPoint(x, y float64) {
	if !b.initialized {
		b.minX, b.minY, b.maxX, b.maxY = x, y, x, y
		b.initialized = true
	} else {
		if x < b.minX {
			b.minX = x
		}
		if x > b.maxX {
			b.maxX = x
		}
		if y < b.minY {
			b.minY = y
		}
		if y > b.maxY {
			b.maxY = y
		}
	}
}

func (b *bboxAccumulator) toBbox(srid int) core.Bbox {
	return core.Bbox{
		Left:   b.minX,
		Bottom: b.minY,
		Right:  b.maxX,
		Top:    b.maxY,
		Srid:   srid,
	}
}

func newWktParser(lexer *wktLexer) *wktParser {
	parser := &wktParser{lexer: lexer}
	parser.nextToken()
	return parser
}

func (p *wktParser) nextToken() {
	p.currentToken = p.lexer.nextToken()
}

func (p *wktParser) expectToken(typ tokenType) error {
	if p.currentToken.typ != typ {
		return fmt.Errorf("expected %v, got %v", typ, p.currentToken.typ)
	}
	return nil
}

func (p *wktParser) parseGeometry() error {
	if p.currentToken.typ == tokenEOF {
		return errors.New("empty input")
	}

	if p.currentToken.typ != tokenWord {
		return errors.New("missing geometry type")
	}

	geometryType := p.currentToken.value
	p.nextToken()

	// Check for EMPTY
	if p.currentToken.typ == tokenWord && p.currentToken.value == "EMPTY" {
		return errors.New("empty geometry not supported")
	}

	switch geometryType {
	case "POINT":
		return p.parsePoint()
	case "LINESTRING":
		return p.parseLineString()
	case "POLYGON":
		return p.parsePolygon()
	case "MULTIPOINT":
		return p.parseMultiPoint()
	case "MULTILINESTRING":
		return p.parseMultiLineString()
	case "MULTIPOLYGON":
		return p.parseMultiPolygon()
	case "GEOMETRYCOLLECTION":
		return p.parseGeometryCollection()
	default:
		return fmt.Errorf("unknown geometry type: %s", geometryType)
	}
}

func (p *wktParser) parsePoint() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	x, y, err := p.parseCoordinate()
	if err != nil {
		return err
	}
	p.bbox.addPoint(x, y)

	// Check for extra coordinates (3D not supported for bounding box)
	if p.currentToken.typ == tokenNumber {
		return errors.New("unexpected coordinate")
	}

	if err := p.expectToken(tokenRParen); err != nil {
		return errors.New("expected closing parenthesis")
	}
	p.nextToken()

	return nil
}

func (p *wktParser) parseLineString() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	pointCount := 0
	for {
		if p.currentToken.typ == tokenRParen {
			break
		}

		if pointCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		x, y, err := p.parseCoordinate()
		if err != nil {
			return err
		}
		p.bbox.addPoint(x, y)
		pointCount++
	}

	if pointCount < 2 {
		return errors.New("linestring requires at least 2 points")
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parsePolygon() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	ringCount := 0
	for {
		if p.currentToken.typ == tokenRParen {
			break
		}
		if p.currentToken.typ == tokenEOF {
			return errors.New("mismatched parentheses")
		}

		if ringCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		if err := p.parseLinearRing(); err != nil {
			return err
		}
		ringCount++
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parseLinearRing() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	var firstX, firstY, lastX, lastY float64
	pointCount := 0

	for {
		if p.currentToken.typ == tokenRParen {
			break
		}
		if p.currentToken.typ == tokenEOF {
			return errors.New("mismatched parentheses")
		}

		if pointCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		x, y, err := p.parseCoordinate()
		if err != nil {
			return err
		}

		if pointCount == 0 {
			firstX, firstY = x, y
		}
		lastX, lastY = x, y
		p.bbox.addPoint(x, y)
		pointCount++
	}

	if pointCount < 4 {
		return errors.New("polygon requires at least 4 points")
	}

	// Check if ring is closed (first and last points should be the same)
	if math.Abs(firstX-lastX) > 1e-9 || math.Abs(firstY-lastY) > 1e-9 {
		return errors.New("polygon ring not closed")
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parseMultiPoint() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	pointCount := 0
	for {
		if p.currentToken.typ == tokenRParen {
			break
		}

		if pointCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		// MULTIPOINT can have two formats: ((x y), (x y)) or (x y, x y)
		if p.currentToken.typ == tokenLParen {
			p.nextToken()
			x, y, err := p.parseCoordinate()
			if err != nil {
				return err
			}
			p.bbox.addPoint(x, y)

			if err := p.expectToken(tokenRParen); err != nil {
				return errors.New("expected closing parenthesis")
			}
			p.nextToken()
		} else {
			x, y, err := p.parseCoordinate()
			if err != nil {
				return err
			}
			p.bbox.addPoint(x, y)
		}
		pointCount++
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parseMultiLineString() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	lineCount := 0
	for {
		if p.currentToken.typ == tokenRParen {
			break
		}

		if lineCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		if err := p.parseLineString(); err != nil {
			return err
		}
		lineCount++
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parseMultiPolygon() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	polygonCount := 0
	for {
		if p.currentToken.typ == tokenRParen {
			break
		}

		if polygonCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		if err := p.parsePolygon(); err != nil {
			return err
		}
		polygonCount++
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parseGeometryCollection() error {
	if err := p.expectToken(tokenLParen); err != nil {
		return errors.New("expected opening parenthesis")
	}
	p.nextToken()

	geomCount := 0
	for {
		if p.currentToken.typ == tokenRParen {
			break
		}

		if geomCount > 0 {
			if err := p.expectToken(tokenComma); err != nil {
				return errors.New("expected comma")
			}
			p.nextToken()
		}

		if err := p.parseGeometry(); err != nil {
			return err
		}
		geomCount++
	}

	p.nextToken() // consume ')'
	return nil
}

func (p *wktParser) parseCoordinate() (float64, float64, error) {
	if p.currentToken.typ != tokenNumber {
		return 0, 0, errors.New("invalid coordinate")
	}

	x, err := strconv.ParseFloat(p.currentToken.value, 64)
	if err != nil {
		return 0, 0, errors.New("invalid coordinate")
	}
	p.nextToken()

	if p.currentToken.typ != tokenNumber {
		return 0, 0, errors.New("incomplete coordinate")
	}

	y, err := strconv.ParseFloat(p.currentToken.value, 64)
	if err != nil {
		return 0, 0, errors.New("invalid coordinate")
	}
	p.nextToken()

	return x, y, nil
}

// parseSrid parses SRID=value; from EWKT format
func (p *wktParser) parseSrid() error {
	// Expect SRID
	if p.currentToken.typ != tokenWord || p.currentToken.value != "SRID" {
		return errors.New("expected SRID")
	}
	p.nextToken()

	// Expect =
	if p.currentToken.typ != tokenEquals {
		return errors.New("expected = after SRID")
	}
	p.nextToken()

	// Expect number
	if p.currentToken.typ != tokenNumber {
		return errors.New("expected SRID number")
	}

	srid, err := strconv.Atoi(p.currentToken.value)
	if err != nil {
		return errors.New("invalid SRID number")
	}
	p.srid = srid
	p.nextToken()

	// Expect ;
	if p.currentToken.typ != tokenSemicolon {
		return errors.New("expected ; after SRID")
	}
	p.nextToken()

	return nil
}

// ParseWkt parses one or more WKT geometries from a reader and returns the combined bounding box
func ParseWkt(r io.Reader) (core.Bbox, error) {
	lexer := newWktLexer(r)
	parser := newWktParser(lexer)

	// Check for empty input first
	if parser.currentToken.typ == tokenEOF {
		return core.Bbox{}, errors.New("empty input")
	}

	// Check for EWKT format with SRID prefix
	if parser.currentToken.typ == tokenWord && parser.currentToken.value == "SRID" {
		if err := parser.parseSrid(); err != nil {
			return core.Bbox{}, err
		}
	}

	// Parse multiple geometries until EOF
	geometryCount := 0
	for parser.currentToken.typ != tokenEOF {
		// Check for error tokens
		if parser.currentToken.typ == tokenError {
			return core.Bbox{}, errors.New(parser.currentToken.value)
		}

		err := parser.parseGeometry()
		if err != nil {
			// Special handling for unknown geometry types in multi-geometry context
			if geometryCount > 0 && strings.Contains(err.Error(), "unknown geometry type") {
				return core.Bbox{}, errors.New("unexpected characters")
			}
			return core.Bbox{}, err
		}
		geometryCount++

		// Check what comes after this geometry
		if parser.currentToken.typ == tokenEOF {
			break
		} else if parser.currentToken.typ == tokenWord {
			// Another geometry follows, continue parsing
			continue
		} else if parser.currentToken.typ == tokenLParen || parser.currentToken.typ == tokenRParen {
			return core.Bbox{}, errors.New("mismatched parentheses")
		} else {
			return core.Bbox{}, errors.New("unexpected characters")
		}
	}

	if geometryCount == 0 || !parser.bbox.initialized {
		return core.Bbox{}, errors.New("empty input")
	}

	return parser.bbox.toBbox(parser.srid), nil
}
