import com.baomidou.mybatisplus.annotation.*;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serializable;
import java.util.Date;

@EqualsAndHashCode(callSuper = false)
@ApiModel(value = "设备表")
@TableName("device_list")
@SuppressWarnings("ALL")
public class DeviceListDO implements Serializable {

    private static final long serialVersionUID = 1L;

    @ApiModelProperty(value = "id")
    @TableId(value = "id", type = IdType.AUTO)
    private Long id;

    @ApiModelProperty(value = "设备id")
    @TableField("device_id")
    private Long deviceId;

    @ApiModelProperty(value = "执行状态 1 未开始 2 执行中 3 执行成功 4 执行失败 5 执行终止")
    @TableField("exec_status")
    private Integer execStatus;

    @ApiModelProperty(value = "执行返回码")
    @TableField("exit_code")
    private Integer exitCode;

    @ApiModelProperty(value = "CPU占用率")
    @TableField("cpu_range")
    private Float CpuRange;

    @ApiModelProperty(value = "XXX目录")
    @TableField("xxx_path")
    private String xxxPath;

    @ApiModelProperty(value = "是否删除 1未删除 2已删除")
    @TableLogic
    private Integer deleted;

    @ApiModelProperty(value = "创建时间")
    @TableField("create_time")
    private Date createTime;

    @ApiModelProperty(value = "修改时间")
    @TableField("update_time")
    private Date updateTime;

}
