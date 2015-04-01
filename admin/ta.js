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
	$('#tax-list tbody').append('<tr id="tax-list-'+t.id+'"><td>'+(new Array(l+1)).join('- ')+t.name+'</td><td>'+t.slug+'</td></tr>');
	if(t.sons) {
		for(var i=0; i<t.sons.length; i++){
			myadd(t.sons[i], l+1);
		}
	}
}

var t = {};
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
			}
		},
		'json'
		);
});

