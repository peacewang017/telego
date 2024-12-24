// notice that this module has been temporarily commented out
// because of building error with go-nvml library in container

// related issues:
// https://github.com/NVIDIA/go-nvml/issues/137
// https://github.com/NVIDIA/go-nvml/issues/49

package app

import (
	// "sync"
	// "time"
	// "github.com/NVIDIA/go-nvml/pkg/nvml"

	"github.com/spf13/cobra"
)

type ModJobInfraExporterStruct struct{}

var ModJobInfraExporter ModJobInfraExporterStruct

// var (
// 	gpuMetricsMutex sync.Mutex

// 	// GPU metrics
// 	gpuUtilization       = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gpu_utilization", Help: "GPU utilization in percentage"}, []string{"index"})
// 	gpuMemoryUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gpu_memory_utilization", Help: "GPU memory utilization in percentage"}, []string{"index"})
// 	gpuPowerUsage        = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gpu_power_usage_mw", Help: "GPU power usage in milliwatts"}, []string{"index"})
// 	gpuTotalMemory       = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gpu_total_memory_bytes", Help: "Total GPU memory in bytes"}, []string{"index"})
// 	gpuUsedMemory        = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "gpu_used_memory_bytes", Help: "Used GPU memory in bytes"}, []string{"index"})
// )

// func init() {
// 	// Register Prometheus metrics
// 	prometheus.MustRegister(gpuUtilization)
// 	prometheus.MustRegister(gpuMemoryUtilization)
// 	prometheus.MustRegister(gpuPowerUsage)
// 	prometheus.MustRegister(gpuTotalMemory)
// 	prometheus.MustRegister(gpuUsedMemory)
// }

// func collectMetrics() {
// 	if ret := nvml.Init(); ret != nvml.SUCCESS {
// 		panic("Failed to initialize NVML")
// 	}
// 	defer nvml.Shutdown()

// 	for {
// 		gpuMetricsMutex.Lock()

// 		deviceCount, _ := nvml.DeviceGetCount()
// 		for i := 0; i < deviceCount; i++ {
// 			device, _ := nvml.DeviceGetHandleByIndex(i)
// 			utilization, _ := device.GetUtilizationRates()
// 			memoryInfo, _ := device.GetMemoryInfo()
// 			powerUsage, _ := device.GetPowerUsage()

// 			// Update Prometheus metrics
// 			gpuUtilization.WithLabelValues(fmt.Sprintf("%d", i)).Set(float64(utilization.Gpu))
// 			gpuMemoryUtilization.WithLabelValues(fmt.Sprintf("%d", i)).Set(float64(utilization.Memory))
// 			gpuPowerUsage.WithLabelValues(fmt.Sprintf("%d", i)).Set(float64(powerUsage))

// 			// Total and Used memory in bytes
// 			gpuTotalMemory.WithLabelValues(fmt.Sprintf("%d", i)).Set(float64(memoryInfo.Total))
// 			gpuUsedMemory.WithLabelValues(fmt.Sprintf("%d", i)).Set(float64(memoryInfo.Used))
// 		}

// 		gpuMetricsMutex.Unlock()
// 		time.Sleep(1 * time.Minute)
// 	}
// }

func (ModJobInfraExporterStruct) Run() {
	// // Start a goroutine to collect metrics
	// go collectMetrics()

	// // Serve metrics
	// http.Handle("/metrics", promhttp.Handler())
	// fmt.Println("Serving metrics on :8080")
	// http.ListenAndServe(":8080", nil)
}

func (ModJobInfraExporterStruct) ParseUpdateInfraExporter() *cobra.Command {
	// job := &ApplyJob{}
	// applyCmd := &cobra.Command{
	// 	Use: "update-infra-exporter",
	// 	// Short: "Apply-related commands",
	// }

	// applyCmd.Run = func(_ *cobra.Command, _ []string) {
	// 	ModJobInfraExporter.Run()
	// }

	// return applyCmd

	return nil
}
