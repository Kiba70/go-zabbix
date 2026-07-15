package zabbix

import (
	"fmt"
	"strconv"
)

// jAction is a private map for the Zabbix API Action object.
// See: https://www.zabbix.com/documentation/2.2/manual/api/reference/action/object
type jAction struct {
	ActionID     string `json:"actionid"`
	EscPeriod    string `json:"esc_period"`
	EvalType     string `json:"evaltype"`
	EventSource  string `json:"eventsource"`
	Name         string `json:"name"`
	DefLongData  string `json:"def_longdata"`
	DefShortData string `json:"def_shortdata"`
	RLongData    string `json:"r_longdata"`
	RShortData   string `json:"r_shortdata"`
	RecoveryMsg  string `json:"recovery_msg"`
	Status       string `json:"status"`
}

// Action returns a native Go Action struct mapped from the given JSON Action
// data.
func (c *jAction) Action() (*Action, error) {
	var err error

	action := &Action{}
	action.ActionID = c.ActionID
	action.Name = c.Name
	action.ProblemMessageSubject = c.DefShortData
	action.ProblemMessageBody = c.DefLongData
	action.RecoveryMessageSubject = c.RShortData
	action.RecoveryMessageBody = c.RShortData
	action.RecoveryMessageEnabled = (c.RecoveryMsg == "1")
	action.Enabled = (c.Status == "0")

	action.StepDuration, err = parseActionStepDuration(c.EscPeriod)
	if err != nil {
		return nil, fmt.Errorf("Error parsing Action Step Duration: %v", err)
	}

	if c.EvalType != "" { // removed in v2.4
		action.EvaluationType, err = strconv.Atoi(c.EvalType)
		if err != nil {
			return nil, fmt.Errorf("Error parsing Action Evaluation Type: %v", err)
		}
	}

	action.EventType, err = strconv.Atoi(c.EventSource)
	if err != nil {
		return nil, fmt.Errorf("Error parsing Action Event Type: %v", err)
	}

	return action, nil
}

func parseActionStepDuration(value string) (int, error) {
	if value == "" {
		return 0, strconv.ErrSyntax
	}

	multiplier := 1
	number := value
	switch value[len(value)-1] {
	case 's':
		number = value[:len(value)-1]
	case 'm':
		multiplier = 60
		number = value[:len(value)-1]
	case 'h':
		multiplier = 60 * 60
		number = value[:len(value)-1]
	case 'd':
		multiplier = 24 * 60 * 60
		number = value[:len(value)-1]
	case 'w':
		multiplier = 7 * 24 * 60 * 60
		number = value[:len(value)-1]
	}

	duration, err := strconv.Atoi(number)
	if err != nil {
		return 0, err
	}

	return duration * multiplier, nil
}
