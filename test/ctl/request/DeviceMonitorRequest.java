
/**
 * 设备监控请求
 */
@Data
@EqualsAndHashCode(callSuper = true)
@ApiModel(value = "设备监控请求")
@SuppressWarnings("ALL")
public class DeviceMonitorRequest extends PageRequest {

    @ApiModelProperty(value = "id")
    private Long id;

    @ApiModelProperty(value = "设备id")
    private Long deviceId;

    @ApiModelProperty(value = "设备名称")
    private String deviceName;

    @ApiModelProperty(value = "请求url")
    private String url;

}
