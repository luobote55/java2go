

/**
 * 设备监控
 */
@Api(tags = "设备监控")
@RestController
@RestWrapper
@RequestMapping("/device/api/monitor")
public class DeviceMonitorController {
    @GetMapping("/get-config")
    @ApiOperation(value = "查询监控配置")
    public DeviceMonitorVO getMonitorConfig(@RequestParam Long deviceId) {
        return deviceMonitorService.getMonitorConfig(deviceId);
    }

    @PostMapping("/set-config")
    @ApiOperation(value = "设置监控配置")
    @EventLog(EventType.UPDATE_MACHINE_MONITOR_CONFIG)
    public DeviceMonitorVO setMonitorConfig(@RequestBody DeviceMonitorRequest request) {
        Valid.notNull(request.getId());
        Valid.notBlank(request.getUrl());
        Valid.notBlank(request.getAccessToken());
        return deviceMonitorService.updateMonitorConfig(request);
    }

}
