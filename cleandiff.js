function getChildren(n, skipMe){
	var r = [];
	for ( ; n; n = n.nextSibling ) 
	   if ( n.nodeType == 1 && n != skipMe)
		  r.push( n );		
	return r;
};

function getChildrenUpTo(n, stopAt){
	var r = [];
	for ( ; n; n = n.nextSibling ) {
		if ( n.nodeType == 1 && n != stopAt) {
			r.push( n );
		} else if ( n == stopAt) {
			break;
		}
	}
	return r;
}


function getSiblings(n) {
	return getChildren(n.parentNode.firstChild, n);
}
function getEarlierSiblings(n){
	return getChildrenUpTo(n.parentNode.firstChild, n);
}

function getPreviousNSiblings(node, count){
	return getEarlierSiblings(node).slice((count * -1));
}

function expandPreviousSiblings(node){
	var predecessors = getPreviousNSiblings(node, 20);
	for (var x = predecessors.length -1; x > -1 ; x--){
		if (predecessors[x].className.indexOf("unchanged") > -1) {
			predecessors[x].className = predecessors[x].className + " expanded";
		}
	}
}

function getAllAnchors(){
	return document.getElementsByClassName("linenum_link");
}

function activateAnchors(){
	var number_anchors = getAllAnchors();
	for (var x = 0; x < number_anchors.length; x++){
		var anchor = number_anchors[x]
		anchor.addEventListener("click",
			function(){expandPreviousSiblings(this.parentNode.parentNode)}, false);
	}
}
