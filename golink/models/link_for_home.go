package models

import (
    "github.com/QLeelulu/goku"
	"github.com/QLeelulu/ohlala/golink/utils"
    "time"
	"fmt"
)

/**
 * 链接推送给网站首页(更新现有数据 )
 */
func link_for_home_update(handleTime time.Time, db *goku.MysqlDB) error {

	sql := `UPDATE tui_link_for_handle H 
		INNER JOIN tui_link_for_home T ON H.insert_time<=? AND H.data_type=2 AND T.link_id=H.link_id 
		INNER JOIN link L ON L.id=H.link_id 
		SET T.score=CASE T.data_type WHEN 2 THEN L.reddit_score -- 热门 
		WHEN 3 THEN ABS(L.vote_up-L.vote_down) -- 热议 
		ELSE L.vote_up-L.vote_down -- 投票 
		END, 
		T.vote_add_score=CASE T.data_type WHEN 3 THEN (L.vote_up+L.vote_down) -- 热议 
		ELSE 0 
		END;`

	_, err := db.Query(sql, handleTime)

	return err
}
/**
 * 链接推送给网站首页(热门)
 */
func link_for_home_top(handleTime time.Time, db *goku.MysqlDB) error {

	sql := `INSERT ignore INTO tui_link_for_home(link_id,create_time,data_type,score,vote_add_score) 
		( 
		SELECT H.link_id,H.create_time,2,L.reddit_score,0 FROM tui_link_for_handle H 
		INNER JOIN link L ON H.insert_time<=? AND L.id=H.link_id 
		); `

	_, err := db.Query(sql, handleTime)

	return err
}
/**
 * 链接推送给网站首页(热议)[3:全部时间；10:这个小时；11:今天；12:这周；13:这个月；14:今年]
 */
func link_for_home_hot(dataType int, handleTime time.Time, db *goku.MysqlDB) error {

	var t time.Time
    switch {
    case dataType == 10:
		t = utils.ThisHour()
    case dataType == 11:
		t = utils.ThisDate()
    case dataType == 12:
		t = utils.ThisWeek()
    case dataType == 13:
		t = utils.ThisMonth()
    case dataType == 14:
		t = utils.ThisYear()
    }
	
	var err error
	if dataType == 3 { //3:全部时间
		sql := `INSERT ignore INTO tui_link_for_home(link_id,create_time,data_type,score,vote_add_score) 
			( 
			SELECT H.link_id,H.create_time,?,ABS(L.vote_up-L.vote_down),L.vote_up+L.vote_down FROM tui_link_for_handle H 
			INNER JOIN link L ON H.insert_time<=? AND L.id=H.link_id 
			); `

		_, err = db.Query(sql, dataType, handleTime)
	} else {
		sql := `INSERT ignore INTO tui_link_for_home(link_id,create_time,data_type,score,vote_add_score) 
		( 
		SELECT H.link_id,H.create_time,?,ABS(L.vote_up-L.vote_down),L.vote_up+L.vote_down FROM tui_link_for_handle H 
		INNER JOIN link L ON H.insert_time<=? AND H.create_time>=? AND L.id=H.link_id 
		); `

		_, err = db.Query(sql, dataType, handleTime, t)
	}


	return err
}

/**
 * 链接推送给网站首页(投票)[投票时间范围: 4:全部时间；5:这个小时；6:今天；7:这周；8:这个月；9:今年]
 */
func link_for_home_vote(dataType int, handleTime time.Time, db *goku.MysqlDB) error {

	var t time.Time
    switch {
    case dataType == 5:
		t = utils.ThisHour()
    case dataType == 6:
		t = utils.ThisDate()
    case dataType == 7:
		t = utils.ThisWeek()
    case dataType == 8:
		t = utils.ThisMonth()
    case dataType == 9:
		t = utils.ThisYear()
    }
	
	var err error
	if dataType == 4 { //4:全部时间
		sql := `INSERT ignore INTO tui_link_for_home(link_id,create_time,data_type,score,vote_add_score) 
			( 
			SELECT H.link_id,H.create_time,?,L.vote_up-L.vote_down,0 FROM tui_link_for_handle H 
			INNER JOIN link L ON H.insert_time<=? AND L.id=H.link_id 
			);`

		_, err = db.Query(sql, dataType, handleTime)
	} else {
		sql := `INSERT ignore INTO tui_link_for_home(link_id,create_time,data_type,score,vote_add_score) 
			( 
			SELECT H.link_id,H.create_time,5,L.vote_up-L.vote_down,0 FROM tui_link_for_handle H 
			INNER JOIN link L ON H.insert_time<=? AND L.create_time>=? AND L.id=H.link_id  
			); `

		_, err = db.Query(sql, dataType, handleTime, t)
	}


	return err
}


/** 删除`tui_link_for_home`
 * 热门, orderName:score DESC,link_id DESC
 * 热议, orderName:score ASC,vote_add_score DESC,link_id DESC
 * 投票, orderName:score DESC,link_id DESC
 */
func del_link_for_home(whereDataType string, orderName string, db *goku.MysqlDB) error {
	
	sql := fmt.Sprintf(`SELECT data_type, tcount - %s AS del_count FROM 
		(SELECT link_id,data_type,COUNT(1) AS tcount FROM tui_link_for_home WHERE %s GROUP BY data_type) T
		WHERE T.tcount>%s;`, LinkMaxCount, whereDataType, LinkMaxCount)

	delSql := `CREATE TEMPORARY TABLE tmp_table 
		( 
		SELECT link_id FROM tui_link_for_home WHERE data_type=%s ORDER BY ` + orderName + ` LIMIT %s,%s 
		); 
		DELETE FROM tui_link_for_home WHERE data_type=%s
		AND link_id IN(SELECT link_id FROM tmp_table); 
		DROP TABLE tmp_table;`
	
	var delCount int64
	var dataType int
	rows, err := db.Query(sql)
	if err == nil {
		for rows.Next() {
			rows.Scan(&dataType, &delCount)
			db.Query(fmt.Sprintf(delSql, dataType, LinkMaxCount, delCount, dataType))
		}
	}

	return err
}


















