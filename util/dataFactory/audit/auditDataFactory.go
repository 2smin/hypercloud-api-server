package audit

import (
	"database/sql"
	"fmt"

	pq "github.com/lib/pq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "tmax"
	DB_NAME     = "postgres"
	HOSTNAME    = "postgres-service.hypercloud5-system.svc"
	PORT        = 5432
)

type Response struct {
	ClusterInfo []ClusterInfo   `json:"clusterInfo"`
	EventList   audit.EventList `json:"eventList"`
	RowsCount   int64           `json:"rowsCount"`
}

type ClusterInfo struct {
	ClusterNamespace string `json:"clusterNamespace"`
	ClusterName      string `json:"clusterName"`
}

var pg_con_info string

func init() {
	pg_con_info = fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		PORT, HOSTNAME, DB_USER, DB_PASSWORD, DB_NAME)
}

func NewNullString(s string) sql.NullString {
	if s == "null" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func Insert(items []audit.Event, clusterName string, clusterNamespace string) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	txn, err := db.Begin()
	if err != nil {
		klog.Error(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("audit", "id", "username", "useragent", "namespace", "apigroup", "apiversion", "resource", "name",
		"stage", "stagetimestamp", "verb", "code", "status", "reason", "message", "clusternamespace", "clustername"))
	if err != nil {
		klog.Error(err)
	}

	for _, event := range items {
		_, err = stmt.Exec(
			event.AuditID,
			event.User.Username,
			event.UserAgent,
			NewNullString(event.ObjectRef.Namespace),
			NewNullString(event.ObjectRef.APIGroup),
			NewNullString(event.ObjectRef.APIVersion),
			event.ObjectRef.Resource,
			event.ObjectRef.Name,
			event.Stage,
			event.StageTimestamp.Time,
			event.Verb,
			event.ResponseStatus.Code,
			event.ResponseStatus.Status,
			event.ResponseStatus.Reason,
			event.ResponseStatus.Message,
			clusterNamespace,
			clusterName)

		if err != nil {
			klog.Error(err)
		}
	}
	res, err := stmt.Exec()
	if err != nil {
		klog.Error(err)
	}

	err = stmt.Close()
	if err != nil {
		klog.Error(err)
	}

	err = txn.Commit()
	if err != nil {
		klog.Error(err)
	}

	if count, err := res.RowsAffected(); err != nil {
		klog.Error(err)
	} else {
		klog.Info("Affected rows: ", count)
	}
}

func Get(query string) Response {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
	}
	defer rows.Close()

	response := Response{}
	clusterInfo := []ClusterInfo{}
	eventList := audit.EventList{}
	var row_count int64
	for rows.Next() {
		var temp_clusternamespace, temp_clustername, temp_namespace, temp_apigroup, temp_apiversion sql.NullString
		cluster := ClusterInfo{}
		event := audit.Event{
			ObjectRef:      &audit.ObjectReference{},
			ResponseStatus: &metav1.Status{},
		}
		err := rows.Scan(
			&event.AuditID,
			&event.User.Username,
			&event.UserAgent,
			&temp_namespace,  //&event.ObjectRef.Namespace,
			&temp_apigroup,   //&event.ObjectRef.APIGroup,
			&temp_apiversion, //&event.ObjectRef.APIVersion,
			&event.ObjectRef.Resource,
			&event.ObjectRef.Name,
			&event.Stage,
			&event.StageTimestamp.Time,
			&event.Verb,
			&event.ResponseStatus.Code,
			&event.ResponseStatus.Status,
			&event.ResponseStatus.Reason,
			&event.ResponseStatus.Message,
			&temp_clusternamespace,
			&temp_clustername,
			&row_count)
		if err != nil {
			rows.Close()
			klog.Error(err)
		}

		if temp_clusternamespace.Valid {
			cluster.ClusterNamespace = temp_clusternamespace.String
		} else {
			cluster.ClusterNamespace = ""
		}

		if temp_clustername.Valid {
			cluster.ClusterNamespace = temp_clustername.String
		} else {
			cluster.ClusterNamespace = ""
		}

		if temp_namespace.Valid {
			event.ObjectRef.Namespace = temp_namespace.String
		} else {
			event.ObjectRef.Namespace = ""
		}

		if temp_apigroup.Valid {
			event.ObjectRef.APIGroup = temp_apigroup.String
		} else {
			event.ObjectRef.APIGroup = ""
		}

		if temp_apiversion.Valid {
			event.ObjectRef.APIVersion = temp_apiversion.String
		} else {
			event.ObjectRef.APIVersion = ""
		}

		clusterInfo = append(clusterInfo, cluster)
		eventList.Items = append(eventList.Items, event)
	}
	eventList.Kind = "EventList"
	eventList.APIVersion = "audit.k8s.io/v1"

	response.ClusterInfo = clusterInfo
	response.EventList = eventList
	response.RowsCount = row_count

	return response
}

func GetMemberList(query string) ([]string, int64) {
	db, err := sql.Open("postgres", pg_con_info)
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		klog.Error(err)
	}
	defer rows.Close()

	// var memberList []string
	memberList := []string{}
	var row_count int64

	for rows.Next() {
		var member string
		err := rows.Scan(
			&member,
			&row_count)
		if err != nil {
			rows.Close()
			klog.Error(err)
		}
		memberList = append(memberList, member)
	}

	return memberList, row_count
}
