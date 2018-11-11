package dmr

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/OmniCar/autobot/vehicle"
	"github.com/mitchellh/hashstructure"
)

// XMLParser represents an XML parser.
type XMLParser struct {
}

// NewXMLParser creates a new XML parser.
func NewXMLParser() *XMLParser {
	return &XMLParser{}
}

// ParseExcerpt parses XML file using XML decoding.
func (p *XMLParser) ParseExcerpt(id int, lines <-chan []string, parsed chan<- vehicle.Vehicle, done chan<- int) {
	var proc, keep int // How many excerpts did we process and keep?
	var stat vehicleStat
	var hash uint64
	for excerpt := range lines {
		if err := xml.Unmarshal([]byte(strings.Join(excerpt, "\n")), &stat); err != nil {
			panic(err) // We _could_ skip it, but it's better to halt execution here.
		}
		if stat.Type == 1 {
			regDate, err := time.Parse("2006-01-02", stat.Info.FirstRegDate[:10])
			if err != nil {
				fmt.Printf("Error: Unable to parse first registration date: %s\n", stat.Info.FirstRegDate)
				continue
			}
			veh := vehicle.Vehicle{
				MetaData:     vehicle.Meta{Source: stat.Info.Source, Country: vehicle.DK, Ident: stat.Ident, LastUpdated: time.Now(), Disabled: false},
				RegNr:        strings.ToUpper(stat.RegNo),
				VIN:          strings.ToUpper(stat.Info.VIN),
				Brand:        vehicle.PrettyBrandName(stat.Info.Designation.BrandTypeName),
				Model:        stat.Info.Designation.Model.Name, // @TODO Title-case model name? Probably difficult.
				FuelType:     vehicle.PrettyFuelType(stat.Info.Engine.Fuel.FuelType),
				FirstRegDate: regDate,
			}
			if hash, err = hashstructure.Hash(veh, nil); err != nil {
				fmt.Printf("Error: Unable to hash Vehicle struct with Ident: %d\n", veh.MetaData.Ident)
				continue
			}
			veh.MetaData.Hash = hash
			parsed <- veh
			keep++
		}
		proc++
	}
	done <- id
}
