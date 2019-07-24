 var idChartUpdater = '';
var periodChartUpdating = 1000;

var myChart;
var dataChart = [];

var isPause = false;

var chartData;


var configChart = {
    type: 'line',
    data: {
        labels: [],
        datasets: []
    },
    options: {
        responsive: true,
        scales: {
            yAxes: [{
                ticks: {
                    beginAtZero: false
                }
            }]
        }
    }
};

function getValue(input) {
    if ( Array.isArray(input) ) {
        return input[0];
    }
    else {
        if (input[input.length-1] == ' ') {
            return input.slice(0, -1);
        }
        else {
            return input;
        }
    }
}

function showValue(indx) {
    $.getJSON( chartData.url + chartData.data[indx]['name'], function(data){
        for ( var j=0; j < data['values'].length; j++  ) {
            if (data['values'][j]['name'] == chartData.data[indx]['value'] ) {
                var inputValue = getValue(data['values'][j]['value']);
                var currentValue;
                if ( inputValue === 'false') { currentValue =0;}
                else if (inputValue === 'true') { currentValue = 1;}
                else { currentValue = inputValue; }

                configChart.data.datasets[indx].data.push(currentValue);
                myChart.update();
                break;
            }
        }
    });
}

function chartUpdating() {
    configChart.data.labels.push(' ');
    for ( var i=0; i < configChart.data.datasets.length; i++) {
        showValue(i);
    }
}

function pauseChartUpdating() {
    if (idChartUpdater != '') {
        stopChartUpdating();
    }
    else {
        startChartUpdating();
    }
}

function startChartUpdating() {
    if (idChartUpdater == '') {
        chartUpdating();
        idChartUpdater = setInterval( chartUpdating, periodChartUpdating);
    }
}

function stopChartUpdating() {
    if (idChartUpdater != '') {
        clearInterval(idChartUpdater);
        idChartUpdater = '';
    }
}

var chartColors = [
    'rgb(255, 99, 132)',
    'rgb(153, 102, 255)',
    'rgb(255, 159, 64)',
    'rgb(255, 205, 86)',
    'rgb(55, 199, 32)',
    'rgb(25, 29, 132)',
    'rgb(255, 199, 132)',
    'rgb(75, 192, 192)',
    'rgb(54, 162, 235)',
    'rgb(201, 203, 207)'
];


function showChart() {
    var ctx = document.getElementById('myChart')
    myChart = new Chart(ctx, configChart);

    for (var i=0; i < chartData.data.length; i++ ) {
		var newColor = chartColors[i];
        var ds = {
            label: chartData.data[i]['name'] + ":" + chartData.data[i]['value'],
            lineTension: 0,
            backgroundColor: newColor,
            borderColor: newColor,
            data: [],
            fill: false
        };
        configChart.data.datasets.push(ds);
    }
    myChart.update();
}

$(document).ready(function(){
    const urlParams = new URLSearchParams(window.location.search);
    const data = urlParams.get('data');
    chartData = JSON.parse(data);

    showChart();
    startChartUpdating();
    isPause = false;

    $('#btn-pause').click(function() {
        if ( !isPause ) {
            stopChartUpdating();
            $(this).addClass('active');

            isPause = true;
        }
        else {
            startChartUpdating();
            isPause = false;
            $(this).removeClass('active');
        }

    });
});