package utils

import "errors"

type DataMatcherWithState struct {
	data_to_match     []byte
	prev_matched_data []byte
}

func NewDataMatcherWithState(data_to_match []byte) DataMatcherWithState {
	return DataMatcherWithState{data_to_match, []byte{}}
}

func (matcher *DataMatcherWithState) Match(data []byte) (int, error) {
	err := errors.New("NOT FOUND")
	result_position := 0
	concat_data := append(matcher.prev_matched_data, data...)
break_out:
	for i := 0; i < len(concat_data); i++ {
		position := 0
		for j := i; j < len(concat_data); j++ {
			if concat_data[j] != matcher.data_to_match[position] {
				break
			}
			position++
			if position == len(matcher.data_to_match) {
				result_position = j - len(matcher.prev_matched_data) - len(matcher.data_to_match) + 1
				err = nil
				break break_out
			}
			if (j + 1) == len(concat_data) {
				matcher.prev_matched_data = concat_data[len(concat_data)-position:]
				break break_out
			}
		}
	}
	return result_position, err
}
