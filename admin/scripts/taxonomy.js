function myloop(t,p,tt) {
	for(var i=0; i<tt.length; i++){
		if(tt[i].parent==p){
			t.sons = t.sons || [];
			t.sons.push(tt[i]);
			myloop(tt[i], tt[i].id, tt);
		}
	}
}

function myadd(t, l) {
	$('#parent').append('<option lv="'+l+'" id="tax-id-'+t.id+'" value="'+t.id+'" class="tax-lv-'+l+'">'+t.name+'</option>');
	$('#tax-list tbody').append(
		'<tr id="tax-list-'+t.id+'">' +
			'<td class="col-name">' +
				'<div class="tax-name">'+(new Array(l+1)).join('- ')+'<span class="real">'+t.name+'</span></div>' +
				'<div class="tax-actions">' +
					'<a class="edit" href="#">编辑</a>' +
					' | ' +
					'<a class="delete" href="#">删除</a>' +
				'</div>' +
			'</td>'  +
			'<td class="col-slug">' + 
				t.slug + 
			'</td>' +
		'</tr>');
	if(t.sons) {
		for(var i=0; i<t.sons.length; i++){
			myadd(t.sons[i], l+1);
		}
	}
}

var t = {sons:[]};
myloop(t, 0, taxes);

for(var i=0; i<t.sons.length; i++){
	myadd(t.sons[i], 0);
}

$('#add-new-tax').click(function() {
	var newid = 0;
	$.post('taxonomy.php',
		$('#new-tax-form').serialize(),
		function(data) {
			if(data.errno === 'success'){
				var parent = $('#parent')[0];
				var selopt = parent.options[parent.selectedIndex];
				var lv = parseInt(selopt.getAttribute('lv'));
				var ids = selopt.getAttribute('id');
				var id = ids.match(/tax-id-(\d+)/)[1];
				if(id != 0) lv+=1;

				var tax = data.tax[0];
				$('<option lv="'+lv+'" value="'+tax.id+'" id="tax-id-'+tax.id+'" class="tax-lv-'+lv+'">'+tax.name+'</option>').insertAfter('#tax-id-'+id);
				$('#tax-list-'+tax.parent).after('<tr id="tax-list-'+tax.id+'"><td>'+(new Array(lv+1)).join('- ')+tax.name+'</td><td>'+tax.slug+'</td></tr>');
			} else {
				alert(data.error);
			}
		},
		'json'
		);
});


$('#tax-list').click(function(e) {
	var cls = e.target.className;
	if(cls == 'edit' || cls == 'delete') {
		// 取得当前操作所在的行
		var tr = e.target.parentNode.parentNode.parentNode;
		var trid = tr.id.replace(/tax-list-(\d+)/, '$1');

		if(cls == 'edit') {
			// 置表单域
			$('#edit-tax input[name="do"]').val(cls === 'edit' ? 'update' : 'delete');
			$('#edit-tax input[name="id"]').val(tr.id.replace(/tax-list-(\d+)/, '$1'));
			$('#edit-tax input[name="name"]').val($('#'+tr.id+' .col-name .tax-name .real').text());
			$('#edit-tax input[name="slug"]').val($('#'+tr.id+' .col-slug').text());

			// 设置显隐
			$('#'+$('#edit-tax')[0].previousSibling.id).show();
			$('#'+tr.id).hide();
			$('#edit-tax').insertAfter($('#'+tr.id)).fadeIn();
		} else if(cls == 'delete') {
			$.post(
				'/admin/taxonomy.php',
                {
                    do: 'delete',
                    id: trid,
                },
				function(data) {
					if(data.errno == 'success') {
						location.href = location.href;
					} else {
						alert(data.error);
					}
				},
				'json'
				);
		}

		e.preventDefault();
	}
});

$('#edit-tax .cancel').click(function(e) {
	$('#'+$('#edit-tax')[0].previousSibling.id).show();
	$('#edit-tax').hide();
	e.preventDefault();
});

$('#edit-tax .submit').click(function(e) {
	$.post(
		'/admin/taxonomy.php',
		$('#edit-tax form').serialize(),
		function(data) {
			if(data.errno == 'success') {
				var trid = '#tax-list-' + $('#edit-tax input[name="id"]').val();
				$(trid + ' .col-name .tax-name .real').text($('#edit-tax input[name="name"]').val());
				$(trid + ' .col-slug').text($('#edit-tax input[name="slug"]').val());

				$('#'+$('#edit-tax')[0].previousSibling.id).show();
				$('#edit-tax').hide();
			} else {
				alert(data.error);
			}
		},
		'json'
		);
	e.preventDefault();
});

