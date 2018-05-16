/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package nodeinfo

const TEMPLATE_PAGE = `
<html>
<head>
<meta http-equiv="refresh" content="3">
<title>Node Information</title>
<style type="text/css">
	a:link {color: #FCFCFC}
	a:visited {color: #FCFCFC}
	a:hover {color: #FCFCFC}
	a:active {color: #FCFCFC}
	body {background:#212124; color:#F8F8FF; font-size:20px;}
	td.bh {color: #00FF00}
	td.pk {word-break: break-all; overflow: hidden}
	table.bd {border: 1px solid #111111; font-size:20px;}
	table.bt {border: 1px solid #111111; font-size:25px;}
	table.font {font-size:20px;}
	a.site {cursor:hand; text-decoration:none;}
</style>
</head>

<body>
<center>
<br><br><br>

<table class="bt" width="80%">
	<tr><th>Node Information</th></tr>
</table>
<br>

<table class="bd" width="80%">
<tr>
<td width="20%" >
	<table class="font" width="100%">
	<tr><th>BlockHeight</th></tr>
	<tr><td align="center"><b><font size="40px">{{.BlockHeight}}</font></b></td></tr>
	</table>
</td>
<td width="80%">
	<table class="font" width="100%">
	<tr><td colspan="1" width="25%">Node Version:</td><td width="25%">{{.NodeVersion}}</td><td width="25%">NodeID:</td><td width="25%">{{.NodeId}}</td></tr>
	<tr><td width="25%">NodeType:</td><td width="25%">{{.NodeType}}</td><td width="25%">NodePort:</td><td width="25%">{{.NodePort}}</td></tr>
	<tr><td width="25%">HttpRestPort:</td><td width="25%">{{.HttpRestPort}}</td><td width="25%">HttpWsPort:</td><td width="25%">{{.HttpWsPort}}</td></tr>
	<tr><td width="25%">HttpJsonPort:</td><td width="25%">{{.HttpJsonPort}}</td><td width="25%">HttpLocalPort:</td><td width="25%">{{.HttpLocalPort}}</td></tr>
	</table>
</td>
</tr>
</table>
<br><br><br><br>

<table class="bt" width="80%">
	<tr><th>Neighbors Information</th></tr>
</table>
<br>

<table class="bd" width="80%">
<tr>
<td width="20%" >
	<table class="font" width="100%">
	<tr><th>Neighbor Count</th></tr>
	<tr><td align="center"><b><font size="40px">{{.NeighborCnt}}</font></b></td></tr>
	</table>
</td>
<td width="80%">
	<table class="font" width="100%">
	<tr><th>Neighbor IP</th><th>Neighbor Id</th><th>Neighbor Type</th></tr>
	{{range .Neighbors}}
	{{if .HttpInfoStart}}
	<tr><td align="center">{{.NgbAddr}}</td><td align="center"><a href="http://{{.HttpInfoAddr}}/info" style="cursor:hand">{{.NgbId}}</a></td><td align="center">{{.NgbType}}</td></tr>
	{{else}}
	<tr><td align="center">{{.NgbAddr}}</td><td align="center">{{.NgbId}}</td><td align="center">{{.NgbType}}</td></tr>
	{{end}}
	{{end}}
	</table>
</td>
</tr>
</table>
<br><br><br><br><br><br>

<table class="font" border="0" width="80%">
	<tr>
	<td width="26%" align="center"><a href="https://ont.io" class="site">site : https://ont.io</a></td>
	</tr>
</table>
<br><br><br><br>

</center>
</body>
</html>
`
