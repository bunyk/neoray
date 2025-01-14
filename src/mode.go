package main

type ModeInfo struct {
	cursor_shape    string
	cell_percentage int
	blinkwait       int
	blinkon         int
	blinkoff        int
	attr_id         int
	attr_id_lm      int
	short_name      string
	name            string
}

type Mode struct {
	cursor_style_enabled bool
	mode_infos           []ModeInfo
	current_mode_name    string
	current_mode         int
}

func CreateMode() Mode {
	return Mode{}
}

func (mode *Mode) Current() ModeInfo {
	if mode.current_mode < len(mode.mode_infos) {
		return mode.mode_infos[mode.current_mode]
	}
	return ModeInfo{}
}

func (mode *Mode) Clear() {
	mode.mode_infos = []ModeInfo{}
}

func (mode *Mode) Add(info ModeInfo) {
	mode.mode_infos = append(mode.mode_infos, info)
}
