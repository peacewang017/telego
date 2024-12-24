package app

import "github.com/thoas/go-funk"

var multiSelectTag = "multi_select"

var selectedTag = "selected"

var exeInstallTag = "exe_install"

var exeApplyTag = "exe_apply"

func (i *MenuItem) isMultiSelect() bool {
	res := false
	for _, tag := range i.SpecTags {
		if tag == multiSelectTag {
			res = true
			break
		}
	}
	return res
}

func (i *MenuItem) setMultiSelectTag() *MenuItem {
	if i.isMultiSelect() {
		return i
	}
	i.SpecTags = append(i.SpecTags, multiSelectTag)
	return i
}

func (i *MenuItem) unsetMultiSelectTag() *MenuItem {
	i.SpecTags = funk.Filter(i.SpecTags, func(tag string) bool { return tag != multiSelectTag }).([]string)
	return i
}

func (i *MenuItem) isSelected() bool {
	res := false
	for _, tag := range i.SpecTags {
		if tag == selectedTag {
			res = true
			break
		}
	}
	return res
}

func (i *MenuItem) setSelectedTag() *MenuItem {
	if i.isSelected() {
		return i
	}
	i.SpecTags = append(i.SpecTags, selectedTag)
	return i
}

func (i *MenuItem) unsetSelectedTag() *MenuItem {
	i.SpecTags = funk.Filter(i.SpecTags, func(tag string) bool { return tag != selectedTag }).([]string)
	return i
}

func (i *MenuItem) isExeInstall() bool {
	res := false
	for _, tag := range i.SpecTags {
		if tag == exeInstallTag {
			res = true
			break
		}
	}
	return res
}

func (i *MenuItem) setExeInstallTag() *MenuItem {
	if i.isExeInstall() {
		return i
	}
	i.SpecTags = append(i.SpecTags, exeInstallTag)
	return i
}

func (i *MenuItem) unsetExeInstallTag() *MenuItem {
	i.SpecTags = funk.Filter(i.SpecTags, func(tag string) bool { return tag != exeInstallTag }).([]string)
	return i
}

func (i *MenuItem) isExeApply() bool {
	res := false
	for _, tag := range i.SpecTags {
		if tag == exeApplyTag {
			res = true
			break
		}
	}
	return res
}

func (i *MenuItem) setExeApplyTag() *MenuItem {
	if i.isExeApply() {
		return i
	}
	i.SpecTags = append(i.SpecTags, exeApplyTag)
	return i
}

func (i *MenuItem) unsetExeApplyTag() *MenuItem {
	i.SpecTags = funk.Filter(i.SpecTags, func(tag string) bool { return tag != exeApplyTag }).([]string)
	return i
}
