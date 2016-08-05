var MSG_OK = 0;
var MSG_ERR = -1;
var MSG_REDIRECT = -2;
	
jQuery(function($){
		
	//ajax提交
	$('.ajax-form').on('submit', function() {
		var url = $('.ajax-form').attr('action') + "?_t=" + Math.random();
		$('.alert').addClass('hide');
		$('button[type="submit"]').attr('disabled', true);
		if ($('#editor')) {
			$("input[name='editor_content']").val($('#editor').html());
		}
		$.post(url, $(".ajax-form").serialize(), function (out) {
			if (out.status == MSG_OK) { // 成功
				if (out.redirect != "") {
					window.location.href = out.redirect;
				} else {
					window.location.reload();
				}
			} else if (out.status == MSG_REDIRECT) {
				window.location.href = out.redirect;
			} else if (out.status == MSG_ERR) {
				if ($('.alert')) {
					$('.alert').removeClass('hide');
					$('.alert').html(out.msg);
				} else {
					alert(out.msg);
				}
				$('button[type="submit"]').removeAttr('disabled');
			}
		});
		return false;
	});
	
    $.datepicker.regional['zh-CN'] = {
        closeText: '关闭',
        prevText: '<上月',
        nextText: '下月>',
        currentText: '今天',
        monthNames: ['一月','二月','三月','四月','五月','六月',
            '七月','八月','九月','十月','十一月','十二月'],
        monthNamesShort: ['一','二','三','四','五','六',
            '七','八','九','十','十一','十二'],
        dayNames: ['星期日','星期一','星期二','星期三','星期四','星期五','星期六'],
        dayNamesShort: ['周日','周一','周二','周三','周四','周五','周六'],
        dayNamesMin: ['日','一','二','三','四','五','六'],
        weekHeader: '周',
        dateFormat: 'yy-mm-dd',
        firstDay: 1,
        isRTL: false,
        showMonthAfterYear: true,
        yearSuffix: '年'};
    $.datepicker.setDefaults($.datepicker.regional['zh-CN']);


    /** 日历控件 **/
    $( "#start_date, #end_date" ).datepicker({
        showOtherMonths: true,
        selectOtherMonths: false,
        dateFormat: 'yy-mm-dd',
        changeYear:true,
        changeMonth:true
    });

    $(function () {
        $('[data-toggle="tooltip"]').tooltip()
    })

    /*$('.delete').click(function () {
        return confirm('确定要删除这条记录吗？');
    });*/
    $.widget("ui.dialog", $.extend({}, $.ui.dialog.prototype, {
        _title: function(title) {
            var $title = this.options.title || '&nbsp;'
            if( ("title_html" in this.options) && this.options.title_html == true )
                title.html($title);
            else title.text($title);
        }
    }));
    $( ".delete_confirm" ).on('click', function(e) {
        var del_url = $(this).attr('href');
        e.preventDefault();
        $( "#dialog-confirm" ).removeClass('hide').dialog({
            resizable: false,
            width: '320',
            modal: true,
            title: "<div class='widget-header'><h4 class='smaller'><i class='ace-icon fa fa-exclamation-triangle red'></i> 删除确认</h4></div>",
            title_html: true,
            buttons: [
                {
                    html: "<i class='ace-icon fa fa-trash-o bigger-110'></i>&nbsp; 确认",
                    "class" : "btn btn-danger btn-sm",
                    click: function() {
                        $( this).dialog('close');
                        window.location.href = del_url;
                    }
                }
                ,
                {
                    html: "<i class='ace-icon fa fa-times bigger-110'></i>&nbsp; 取消",
                    "class" : "btn btn-sm",
                    click: function() {
                        $( this ).dialog( "close" );
                    }
                }
            ]
        });
    });


});