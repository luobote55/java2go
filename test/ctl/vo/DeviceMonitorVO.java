
/**
 * 设备监控响应
 */
@Data
@ApiModel(value = "设备监控响应")
@SuppressWarnings("ALL")
public class DeviceMonitorVO {

    @ApiModelProperty(value = "id")
    private Long id;

    @ApiModelProperty(value = "设备名称")
    private String deviceName;

    @ApiModelProperty(value = "监控状态 1未安装 2安装中 3未运行 4运行中")
    private Integer status;

    @ApiModelProperty("机器监控 url")
    private String url;
}
