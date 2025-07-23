
function hostmap_ui(o){
    console.log(o);
    var hH = '<option value="" >Select...</option>';
    var hD = '<option value="" >Select...</option>';
    for (const [key, value] of Object.entries(o)) {
        hH += '<option value="'+key+'">'+key+'</option>';
        value.forEach((valueD) => {
            hD += '<option value="'+valueD+'">'+valueD+'</option>';
        });
    }
    document.getElementById('select-host').innerHTML = hH;
    document.getElementById('select-date').innerHTML = hD;
}



function hostmap_ui_update(){
    document.getElementById('paila_log_content').innerHTML = '';
    var h = document.getElementById('select-host').value;
    var d = document.getElementById('select-date').value;
    

    if(h=="" || d==""){
        return;
    }

    window.location.hash = '#?host='+encodeURIComponent(h)+'&date='+encodeURIComponent(d)+''

    const url = '/report-data?host='+encodeURIComponent(h)+'&date='+encodeURIComponent(d)+''; // Replace with your actual API endpoint

    // Use the Fetch API to make a GET request
    fetch(url)
        .then(response => {
            // Check if the request was successful (status code 200-299)
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            // Parse the JSON response
            return response.json();
        })
        .then(data => {
            // Handle the parsed JSON data
            //console.log('Received JSON data:', data);

            dataReport = escapeHtml(data.report);
            if(data.report==""){
                dataReport="<div class=\"no-report-message-generate\">No Report Generated<br /><br /><button onclick=\"hostmap_ui_generate()\">Generate Now</button></div>"
            } else {
                dataReport=markdownToHtml(dataReport)+"<div style=\"text-align:right;padding:42px;\"><button onclick=\"hostmap_ui_generate()\">Regenerate Report</button></div>"
            }




            document.getElementById('paila_log_content').innerHTML = '<div><span>'+data.host+' : '+data.date+'</span> &nbsp; '+
                '<span id="paila_content_tab_report" onclick="hostmap_ui_tab(\'report\')">Report</span> &nbsp;'+
                '<span id="paila_content_tab_logs" onclick="hostmap_ui_tab(\'logs\')">Logs</span> &nbsp; '+
                '<span id="paila_content_tab_specs" onclick="hostmap_ui_tab(\'specs\')">Specs</span></div>'+
                '<div id="paila_log_tab_content">'+
                '<div id="paila_content_report">'+dataReport+'</div>'+
                '<div id="paila_content_logs"><pre>'+escapeHtml(data.logs)+'</pre></div>'+
                '<div id="paila_content_specs"><pre>'+escapeHtml(data.specs)+'</pre></div>'+
                '</div>';
            hostmap_ui_tab('report');
        })
        .catch(error => {
            // Handle any errors that occurred during fetch 
            console.error('Error fetching data:', error);
            document.getElementById('paila_log_content').innerHTML = 'Error fetching data';
        });


    
}



function escapeHtml(text) {
  var map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}




function hostmap_ui_tab(t){
    document.getElementById('paila_content_logs').style.display = 'none';
    document.getElementById('paila_content_specs').style.display = 'none';
    document.getElementById('paila_content_report').style.display = 'none';

    document.getElementById('paila_content_tab_logs').style.opacity = '0.42';
    document.getElementById('paila_content_tab_specs').style.opacity = '0.42';
    document.getElementById('paila_content_tab_report').style.opacity = '0.42';

    if(t =="logs"){
        document.getElementById('paila_content_logs').style.display = 'block';
        document.getElementById('paila_content_tab_logs').style.opacity = '1';
    }
    if(t =="specs"){
        document.getElementById('paila_content_specs').style.display = 'block';
        document.getElementById('paila_content_tab_specs').style.opacity = '1';
    }
    if(t =="report"){
        document.getElementById('paila_content_report').style.display = 'block';
        document.getElementById('paila_content_tab_report').style.opacity = '1';
    }

    /*
    switch (t) {
        case "logs": 
            document.getElementById('paila-logs').style.display = 'block';
            document.getElementById('paila-content-tab-logs').style.opacity = '1';
        case "specs":
            document.getElementById('paila-specs').style.display = 'block';
            document.getElementById('paila-content-tab-specs').style.opacity = '1';
        case "report":
            document.getElementById('paila-report').style.display = 'block';
            document.getElementById('paila-content-tab-report').style.opacity = '1';
    }
    */
}



async function hostmap_ui_generate(){
    var h = document.getElementById('select-host').value;
    var d = document.getElementById('select-date').value;
    if(h=="" || d==""){
        return;
    }
    document.getElementById('paila_content_report').innerHTML = "<div class=\"no-report-message-generate\">&nbsp;<br /><br />Generating ...</div>";

    //var u = 'http://localhost/report-generate?host='+encodeURIComponent(h)+'&date='+encodeURIComponent(d)+''
    var u = '/report-generate?host='+encodeURIComponent(h)+'&date='+encodeURIComponent(d)+''

    try {
        const response = await fetch(u);
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`);
            //console.log(response);
            // reload get
            document.getElementById('paila_content_report').innerHTML = '<pre>!OK</pre>';
        }
        document.getElementById('paila_content_report').innerHTML = '<pre>OK</pre>';
        //if(response) // "success":"1"
        //const json = await response.json();
        console.log(response);
        //document.getElementById('paila_content_report').innerHTML = '<pre>'+escapeHtml(response)+'</pre>';
        hostmap_ui_update()
    } catch (error) {
        document.getElementById('paila_content_report').innerHTML = '<pre>'+escapeHtml(error.message)+'</pre>';
        setTimeout(hostmap_ui_update,1500);
    }
    
}



function markdownToHtml(markdown) {
  let html = markdown;

  // Convert headings
  html = html.replace(/^###### (.*$)/gim, '<h6>$1</h6>');
  html = html.replace(/^##### (.*$)/gim, '<h5>$1</h5>');
  html = html.replace(/^#### (.*$)/gim, '<h4>$1</h4>');
  html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
  html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
  html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');

  // Convert bold text
  html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/__(.*?)__/g, '<strong>$1</strong>');

  // convert --- to horizontal rule
  html = html.replace(/^-{3,}$/gm, '<hr>');

  // convert newlines to <br>
  html = html.replace(/\n/g, '<br>\n');

  // convert lists to <ul> and <li>
  html = html.replace(/^\s*[-*]\s+(.*)$/gm, '<li>$1</li>');
  html = html.replace(/(<li>.*<\/li>)/g, '<ul>$1</ul>');
    



  // Add more rules for other Markdown elements (italics, lists, links, etc.)

  return html;
}
