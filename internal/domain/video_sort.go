package domain

type VideoSort string

const (
	VideoSortNewest VideoSort = "newest"
	VideoSortViews  VideoSort = "views"
	VideoSortRating VideoSort = "rating"
)
