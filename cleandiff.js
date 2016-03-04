function getChildren(node, skipMe){
	var r = [];
	for ( ; node; node = node.nextSibling ) 
	   if ( node.nodeType == 1 && node != skipMe)
		  r.push( node );		
	return r;
};

function getChildrenUpTo(node, stopAt){
	var r = [];
	for ( ; node; node = node.nextSibling ) {
		if ( node.nodeType == 1 && node != stopAt) {
			if (node.classList[0] == "line") {
			// ^^ only to compensate for bug in bash version's source html
				r.push( node );
			}
		} else if ( node == stopAt) {
			break;
		}
	}
	return r;
}
function getLaterNSiblings(node, n){
	var r = [];
	for ( ; node; node = node.nextSibling ) {
		if ( r.length < n && node.nodeType == 1) {
			// nodeType 3 is Text 1 is element. 
			r.push( node )
		} else if (r.length == n) {
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

function expandPreviousSiblings(node, how_many){
	var predecessors = getPreviousNSiblings(node, how_many);
	for (var x = predecessors.length -1; x > -1 ; x--){
		expandNodeIfUnchanged(predecessors[x]);
	}
}
function expandLaterSiblings(node, how_many){
	var antecessors = getLaterNSiblings(node, how_many);
	for (var x = 0; x < antecessors.length; x++){
		expandNodeIfUnchanged( antecessors[x])
	}
}
function expandNodeIfUnchanged(node){
	if (typeof(node) != "undefined"){
		if (node.className.indexOf("unchanged") > -1) {
			node.className = node.className + " expanded";
		}
	}
}

function getAllAnchors(){
	return document.getElementsByClassName("linenum_link");
}

function getChangedLineDivs(){
	var all_lines = document.getElementsByClassName("line");
	var changed_lines = [];
	for (var x = 0; x < all_lines.length; x++){
		var node = all_lines[x];
		if (node.className.indexOf("unchanged") == -1) {
			changed_lines.push(node);
		}
	}
	return changed_lines;
}
function addOnClickToAnchors(number_anchors){
	for (var x = 0; x < number_anchors.length; x++){
		var anchor = number_anchors[x]
		anchor.addEventListener("click",
			function(){expandPreviousSiblings(this.parentNode.parentNode, 20)}, false);
	}
}
function expandAroundChangedLines(){
	var changed_lines = getChangedLineDivs();
	for (var x = 0; x < changed_lines.length; x++){
		var node = changed_lines[x]
		expandPreviousSiblings(node, 5);
		expandLaterSiblings(node, 5);
	}
}

function activateAnchors(){
	var number_anchors = getAllAnchors();
	addOnClickToAnchors(number_anchors);
	expandAroundChangedLines()
}

Array.prototype.first = function(){
	if (this.length > 1){
		return this[0];
	}
	return null;
}

Array.prototype.rest = function(){
	if (this.length > 1){
		return this.slice(1)
	}
	return new Array();
}

