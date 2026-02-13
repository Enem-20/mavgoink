package message

import "github.com/snksoft/crc"

type CRC struct {
	crcCalculator     *crc.Table
	MCRF4XXparameters *crc.Parameters
	Raw               uint64 `json:"crc"`
}

func newCRC() *CRC {
	MCRF4XXparameters := &crc.Parameters{
		Width:      16,
		Polynomial: 0x1021,
		Init:       0xFFFF,
		ReflectIn:  true,
		ReflectOut: true,
		FinalXor:   0xFFFF,
	}

	return &CRC{crcCalculator: crc.NewTable(MCRF4XXparameters), MCRF4XXparameters: MCRF4XXparameters}
}

func (c *CRC) Reset() {
	c.Raw = c.crcCalculator.InitCrc()
}

func (c *CRC) CalculateByte(value byte) {
	c.Raw = c.crcCalculator.UpdateCrc(c.Raw, []byte{value})
}

func (c *CRC) Calculate(data []byte) {
	c.Raw = c.crcCalculator.UpdateCrc(c.Raw, data)
}

func (c *CRC) GetCRC() uint16 {
	return c.crcCalculator.CRC16(c.Raw)
}
