package routers
																																																							
import (
	"github.com/gin-gonic/gin"
	jwt "qzkvm/src/middleware"
	setting2 "qzkvm/src/pkg/setting"
	"qzkvm/src/qzcloud"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())

	r.Use(gin.Recovery())

	gin.SetMode(setting2.RunMode)
	r.Use(jwt.JWT())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "test",
		})
	})
    apiv1:=r.Group("/api/v1")
    apiv1.POST("/create_kvm",qzcloud.CreateKvm)
	apiv1.POST("/update_kvm",qzcloud.UpdateKvm)
	apiv1.POST("/remove_kvm",qzcloud.RemoveKvm)
	apiv1.POST("/reinstall_kvm",qzcloud.ReinstallKvm)
	apiv1.POST("/update_system_password",qzcloud.UpdateSystemPassword)
	apiv1.POST("/create_snapshot_kvm",qzcloud.CreateSnapshotkvm)
	apiv1.POST("/restore_snapshot_kvm",qzcloud.RestoreSnapshotKvm)
	apiv1.POST("/remove_snapshot_kvm",qzcloud.RemoveSnapshotKvm)
	apiv1.POST("/update_ip_kvm",qzcloud.UpdateIPKvm)
	apiv1.POST("/add_firewall_kvm",qzcloud.AddNWFilterKvm)
	apiv1.POST("/remove_firewall_kvm",qzcloud.RemoveNWFilterKvm)
	apiv1.POST("/update_iso_kvm",qzcloud.UpdateIsoKvm)
	apiv1.POST("/get_screenshot_kvm",qzcloud.GetScreenshotKvm)
	apiv1.POST("/create_backup_kvm",qzcloud.CreateBackupKvm)
	apiv1.POST("/restore_backup_kvm",qzcloud.RestoreBackupKvm)
	apiv1.POST("/remove_backup_kvm",qzcloud.RemoveBackupKvm)
	apiv1.POST("/monitor_kvm",qzcloud.MonitorKvm)
	apiv1.POST("/boot_order",qzcloud.BootOrderKvm)
	apiv1.POST("/get_network_flow_kvm",qzcloud.GetNetworkFlowKvm)
	apiv1.POST("/get_status_kvm",qzcloud.GetStatusKvm)
	apiv1.POST("/set_status_kvm",qzcloud.SetStatusKvm)
	apiv1.POST("/get_isolist_kvm",qzcloud.GetIsoListKvm)
	apiv1.POST("/set_network_state",qzcloud.NetworkCloseOrOpenKvm)
	apiv1.POST("/count_flow_kvm",qzcloud.CountFlowKvm)

	return r
}