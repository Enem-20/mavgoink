package system

import (
	"strconv"

	"github.com/Enem-20/mavgoink/component"
	"github.com/Enem-20/mavgoink/message"
)

const (
	MAVLINK_VERSION_1 = 0xFE
	MAVLINK_VERSION_2 = 0xFD
	DEFAULT_SYSTEM_ID = 1
	componentsCount   = 256
	minComponentID    = 1
	maxComponentID    = 255
)

type System struct {
	STX                byte                                  `json:"stx"`
	ID                 byte                                  `json:"id"`
	Name               string                                `json:"name"`
	Components         [componentsCount]*component.Component `json:"components"`
	CurrentComponentID byte                                  `json:"current_component_id"`
	Seq                byte                                  `json:"seq"`
}

func NewSystem(stx byte, id byte, name string) *System {
	system := &System{STX: stx, ID: id, Name: name, Components: [componentsCount]*component.Component{}, Seq: 0}
	system.PushBackDefaultComponent()
	return system
}

func (s *System) getNewIndex() byte {
	if s.CurrentComponentID >= maxComponentID {
		return 0
	}
	return byte(s.CurrentComponentID + 1)
}

func (s *System) PushBackDefaultComponent() {
	defaultComponent := component.NewComponent(s.getNewIndex(), "Default Component "+strconv.Itoa(int(s.getNewIndex())))
	s.Components[s.CurrentComponentID] = defaultComponent
	s.CurrentComponentID++
}

func (s *System) PushBackComponent(comp *component.Component) {
	s.Components[s.CurrentComponentID] = comp
	s.CurrentComponentID++
}

func (s *System) PlaceComponentAtIndex(comp *component.Component, index byte) {
	if index == 0 {
		return
	}
	if index > maxComponentID {
		return
	}
	s.Components[index-1] = comp
}

func (s *System) GetComponentByID(id byte) *component.Component {
	if (id == 0) || (id > maxComponentID) {
		return nil
	}

	return s.Components[id-1]
}

// Slower then GetComponentByID, but more intuitive to use (O(n) vs O(1))
func (s *System) GetComponentByName(name string) *component.Component {
	for _, comp := range s.Components {
		if comp.Name == name {
			return comp
		}
	}
	return nil
}

func (s *System) CreateDefaultMessage(compId byte, msgId uint32) *message.Message {
	return s.CreateMessage(compId, msgId, 0)
}

func (s *System) CreateMessage(compId byte, msgId uint32, payloadCapacity byte) *message.Message {
	s.Seq = s.Seq%255 + 1
	return message.NewMessageFrom(s.STX, payloadCapacity, s.Seq, s.ID, compId, msgId)
}
