package douyin

type DouYinResult struct {
	Url            string  `json:"url"`
	Endpoint       string  `json:"endpoint"`
	TotalTime      float64 `json:"total_time"`
	Status         string  `json:"status"`
	Message        string  `json:"message"`
	Type           string  `json:"type"`
	Platform       string  `json:"platform"`
	AwemeId        string  `json:"aweme_id"`
	OfficialApiUrl struct {
		UserAgent string `json:"User-Agent"`
		ApiUrl    string `json:"api_url"`
	} `json:"official_api_url"`
	Desc       string `json:"desc"`
	CreateTime int    `json:"create_time"`
	Author     struct {
		AvatarThumb struct {
			Height  float64  `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   float64  `json:"width"`
		} `json:"avatar_thumb"`
		CfList          interface{} `json:"cf_list"`
		CloseFriendType int         `json:"close_friend_type"`
		ContactsStatus  int         `json:"contacts_status"`
		ContrailList    interface{} `json:"contrail_list"`
		CoverUrl        []struct {
			Height  float64  `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   float64  `json:"width"`
		} `json:"cover_url"`
		CustomVerify           string      `json:"custom_verify"`
		DataLabelList          interface{} `json:"data_label_list"`
		EndorsementInfoList    interface{} `json:"endorsement_info_list"`
		EnterpriseVerifyReason string      `json:"enterprise_verify_reason"`
		FamiliarVisitorUser    interface{} `json:"familiar_visitor_user"`
		ImRoleIds              interface{} `json:"im_role_ids"`
		IsAdFake               bool        `json:"is_ad_fake"`
		IsBan                  bool        `json:"is_ban"`
		IsBlockedV2            bool        `json:"is_blocked_v2"`
		IsBlockingV2           bool        `json:"is_blocking_v2"`
		Nickname               string      `json:"nickname"`
		NotSeenItemIdList      interface{} `json:"not_seen_item_id_list"`
		NotSeenItemIdListV2    interface{} `json:"not_seen_item_id_list_v2"`
		OfflineInfoList        interface{} `json:"offline_info_list"`
		PersonalTagList        interface{} `json:"personal_tag_list"`
		PreventDownload        bool        `json:"prevent_download"`
		RiskNoticeText         string      `json:"risk_notice_text"`
		SecUid                 string      `json:"sec_uid"`
		ShareInfo              struct {
			ShareDesc      string `json:"share_desc"`
			ShareDescInfo  string `json:"share_desc_info"`
			ShareQrcodeUrl struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"share_qrcode_url"`
			ShareTitle       string `json:"share_title"`
			ShareTitleMyself string `json:"share_title_myself"`
			ShareTitleOther  string `json:"share_title_other"`
			ShareUrl         string `json:"share_url"`
			ShareWeiboDesc   string `json:"share_weibo_desc"`
		} `json:"share_info"`
		ShortId             string      `json:"short_id"`
		Signature           string      `json:"signature"`
		SignatureExtra      interface{} `json:"signature_extra"`
		SpecialFollowStatus int         `json:"special_follow_status"`
		SpecialPeopleLabels interface{} `json:"special_people_labels"`
		Status              int         `json:"status"`
		TextExtra           interface{} `json:"text_extra"`
		TotalFavorited      int         `json:"total_favorited"`
		Uid                 string      `json:"uid"`
		UniqueId            string      `json:"unique_id"`
		UserAge             int         `json:"user_age"`
		UserCanceled        bool        `json:"user_canceled"`
		UserPermissions     interface{} `json:"user_permissions"`
		VerificationType    int         `json:"verification_type"`
	} `json:"author"`
	Music struct {
		Album            string        `json:"album"`
		ArtistUserInfos  interface{}   `json:"artist_user_infos"`
		Artists          []interface{} `json:"artists"`
		AuditionDuration int           `json:"audition_duration"`
		Author           string        `json:"author"`
		AuthorDeleted    bool          `json:"author_deleted"`
		AuthorPosition   interface{}   `json:"author_position"`
		AuthorStatus     int           `json:"author_status"`
		AvatarLarge      struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"avatar_large"`
		AvatarMedium struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"avatar_medium"`
		AvatarThumb struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"avatar_thumb"`
		BindedChallengeId int  `json:"binded_challenge_id"`
		CanBackgroundPlay bool `json:"can_background_play"`
		Climax            struct {
			StartPoint int `json:"start_point"`
		} `json:"climax"`
		CoverColorHsv struct {
			H int `json:"h"`
			S int `json:"s"`
			V int `json:"v"`
		} `json:"cover_color_hsv"`
		CoverHd struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"cover_hd"`
		CoverLarge struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"cover_large"`
		CoverMedium struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"cover_medium"`
		CoverThumb struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"cover_thumb"`
		DmvAutoShow          bool          `json:"dmv_auto_show"`
		Duration             float64       `json:"duration"`
		EndTime              int           `json:"end_time"`
		ExternalSongInfo     []interface{} `json:"external_song_info"`
		Extra                string        `json:"extra"`
		Id                   int64         `json:"id"`
		IdStr                string        `json:"id_str"`
		IsAudioUrlWithCookie bool          `json:"is_audio_url_with_cookie"`
		IsCommerceMusic      bool          `json:"is_commerce_music"`
		IsDelVideo           bool          `json:"is_del_video"`
		IsMatchedMetadata    bool          `json:"is_matched_metadata"`
		IsOriginal           bool          `json:"is_original"`
		IsOriginalSound      bool          `json:"is_original_sound"`
		IsPgc                bool          `json:"is_pgc"`
		IsRestricted         bool          `json:"is_restricted"`
		IsVideoSelfSee       bool          `json:"is_video_self_see"`
		LunaInfo             struct {
			HasCopyright bool `json:"has_copyright"`
			IsLunaUser   bool `json:"is_luna_user"`
		} `json:"luna_info"`
		LyricShortPosition             interface{} `json:"lyric_short_position"`
		Mid                            string      `json:"mid"`
		MusicChartRanks                interface{} `json:"music_chart_ranks"`
		MusicCollectCount              int         `json:"music_collect_count"`
		MusicCoverAtmosphereColorValue string      `json:"music_cover_atmosphere_color_value"`
		MusicianUserInfos              interface{} `json:"musician_user_infos"`
		OfflineDesc                    string      `json:"offline_desc"`
		OwnerHandle                    string      `json:"owner_handle"`
		OwnerId                        string      `json:"owner_id"`
		OwnerNickname                  string      `json:"owner_nickname"`
		PlayUrl                        struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlKey  string   `json:"url_key"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"play_url"`
		Redirect   bool   `json:"redirect"`
		SchemaUrl  string `json:"schema_url"`
		SearchImpr struct {
			EntityId string `json:"entity_id"`
		} `json:"search_impr"`
		SecUid        string `json:"sec_uid"`
		ShootDuration int    `json:"shoot_duration"`
		Song          struct {
			Artists interface{} `json:"artists"`
			Chorus  struct {
				DurationMs int `json:"duration_ms"`
				StartMs    int `json:"start_ms"`
			} `json:"chorus"`
			ChorusV3Infos interface{} `json:"chorus_v3_infos"`
			Id            int64       `json:"id"`
			IdStr         string      `json:"id_str"`
			Title         string      `json:"title"`
		} `json:"song"`
		StrongBeatUrl struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"strong_beat_url"`
		TagList           interface{} `json:"tag_list"`
		Title             string      `json:"title"`
		UnshelveCountries interface{} `json:"unshelve_countries"`
		UserCount         int         `json:"user_count"`
		VideoDuration     int         `json:"video_duration"`
	} `json:"music"`
	Statistics struct {
		AdmireCount  int    `json:"admire_count"`
		AwemeId      string `json:"aweme_id"`
		CollectCount int    `json:"collect_count"`
		CommentCount int    `json:"comment_count"`
		DiggCount    int    `json:"digg_count"`
		PlayCount    int    `json:"play_count"`
		ShareCount   int    `json:"share_count"`
	} `json:"statistics"`
	CoverData struct {
		Cover struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"cover"`
		OriginCover struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"origin_cover"`
		DynamicCover struct {
			Height  int      `json:"height"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
			Width   int      `json:"width"`
		} `json:"dynamic_cover"`
	} `json:"cover_data"`
	Hashtags []struct {
		End         int    `json:"end"`
		HashtagId   string `json:"hashtag_id"`
		HashtagName string `json:"hashtag_name"`
		IsCommerce  bool   `json:"is_commerce"`
		Start       int    `json:"start"`
		Type        int    `json:"type"`
	} `json:"hashtags"`
	VideoData struct {
		WmVideoUrl    string `json:"wm_video_url"`
		WmVideoUrlHQ  string `json:"wm_video_url_HQ"`
		NwmVideoUrl   string `json:"nwm_video_url"`
		NwmVideoUrlHQ string `json:"nwm_video_url_HQ"`
	} `json:"video_data"`
	Images []Image `json:"images"`
}

type Image struct {
	Height  int      `json:"height"`
	Width   int      `json:"width"`
	URLList []string `json:"url_list"`
	URI     string   `json:"uri"`
}
