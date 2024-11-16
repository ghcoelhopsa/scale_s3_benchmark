// static/js/charts.js

let ctx = document.getElementById('uploadsChart').getContext('2d');
let uploadsChart = new Chart(ctx, {
    type: 'line', // Gráfico de linha
    data: {
        labels: [], // Labels de tempo
        datasets: [
            {
                label: 'Sucessos',
                data: [],
                backgroundColor: 'rgba(75, 192, 192, 0.2)',
                borderColor: 'rgba(75, 192, 192, 1)',
                borderWidth: 1,
                fill: false
            },
            {
                label: 'Falhas',
                data: [],
                backgroundColor: 'rgba(255, 99, 132, 0.2)',
                borderColor: 'rgba(255, 99, 132, 1)',
                borderWidth: 1,
                fill: false
            }
        ]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            x: {
                type: 'time',
                time: {
                    unit: 'minute',
                    tooltipFormat: 'HH:mm:ss'
                },
                title: {
                    display: true,
                    text: 'Tempo'
                }
            },
            y: {
                beginAtZero: true,
                precision: 0,
                title: {
                    display: true,
                    text: 'Número de Uploads'
                }
            }
        }
    }
});

// Gráficos adicionais para taxas (Uploads por Segundo e por Hora)
let ctxPerSecond = document.getElementById('uploadsPerSecondChart').getContext('2d');
let uploadsPerSecondChart = new Chart(ctxPerSecond, {
    type: 'line',
    data: {
        labels: [],
        datasets: [{
            label: 'Uploads por Segundo',
            data: [],
            backgroundColor: 'rgba(54, 162, 235, 0.2)',
            borderColor: 'rgba(54, 162, 235, 1)',
            borderWidth: 1,
            fill: false
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            x: {
                type: 'time',
                time: {
                    unit: 'minute',
                    tooltipFormat: 'HH:mm:ss'
                },
                title: {
                    display: true,
                    text: 'Tempo'
                }
            },
            y: {
                beginAtZero: true,
                precision: 0,
                title: {
                    display: true,
                    text: 'Uploads por Segundo'
                }
            }
        }
    }
});

let ctxPerHour = document.getElementById('uploadsPerHourChart').getContext('2d');
let uploadsPerHourChart = new Chart(ctxPerHour, {
    type: 'line',
    data: {
        labels: [],
        datasets: [{
            label: 'Uploads por Hora',
            data: [],
            backgroundColor: 'rgba(255, 206, 86, 0.2)',
            borderColor: 'rgba(255, 206, 86, 1)',
            borderWidth: 1,
            fill: false
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            x: {
                type: 'time',
                time: {
                    unit: 'hour',
                    tooltipFormat: 'HH:mm'
                },
                title: {
                    display: true,
                    text: 'Tempo'
                }
            },
            y: {
                beginAtZero: true,
                precision: 0,
                title: {
                    display: true,
                    text: 'Uploads por Hora'
                }
            }
        }
    }
});

// Histórico para calcular as taxas
let uploadHistory = []; // Array de objetos { timestamp: Date, successes: Number, failures: Number }

// Função para atualizar as estatísticas na página
function updateStats(data) {
    const now = new Date();

    document.getElementById('total-uploads').innerText = data.TotalUploads;
    document.getElementById('successes').innerText = data.Successes;
    document.getElementById('failures').innerText = data.Failures;

    // Atualizar gráfico principal
    uploadsChart.data.labels.push(now);
    uploadsChart.data.datasets[0].data.push(data.Successes);
    uploadsChart.data.datasets[1].data.push(data.Failures);

    const maxDataPoints = 20;
    if (uploadsChart.data.labels.length > maxDataPoints) {
        uploadsChart.data.labels.shift();
        uploadsChart.data.datasets.forEach(dataset => dataset.data.shift());
    }

    uploadsChart.update();

    // Atualizar histórico
    uploadHistory.push({
        timestamp: now,
        successes: data.Successes,
        failures: data.Failures
    });

    // Limpar histórico além de 1 hora
    const oneHourAgo = new Date(now.getTime() - (60 * 60 * 1000));
    uploadHistory = uploadHistory.filter(entry => entry.timestamp >= oneHourAgo);

    // Calcular uploads por segundo e por hora
    calculateUploadRates(now);

    // Atualizar gráficos de taxas
    updateRateCharts(now);
}

// Função para calcular uploads por segundo e por hora
function calculateUploadRates(currentTime) {
    // Calcular uploads por segundo
    let uploadsInLastSecond = 0;
    uploadHistory.forEach(entry => {
        if ((currentTime - entry.timestamp) <= 1000) { // 1 segundo = 1000 ms
            uploadsInLastSecond += entry.successes + entry.failures;
        }
    });

    document.getElementById('uploads-per-second').innerText = uploadsInLastSecond;

    // Calcular uploads por hora
    let uploadsInLastHour = 0;
    uploadHistory.forEach(entry => {
        uploadsInLastHour += entry.successes + entry.failures;
    });

    document.getElementById('uploads-per-hour').innerText = uploadsInLastHour;
}

// Função para atualizar os gráficos de taxas
function updateRateCharts(currentTime) {
    // Atualizar Uploads por Segundo
    uploadsPerSecondChart.data.labels.push(currentTime);
    let uploadsInLastSecond = 0;
    uploadHistory.forEach(entry => {
        if ((currentTime - entry.timestamp) <= 1000) { // 1 segundo
            uploadsInLastSecond += entry.successes + entry.failures;
        }
    });
    uploadsPerSecondChart.data.datasets[0].data.push(uploadsInLastSecond);

    // Limitar pontos no gráfico
    const maxDataPointsPerSecond = 60; // Últimos 60 segundos
    if (uploadsPerSecondChart.data.labels.length > maxDataPointsPerSecond) {
        uploadsPerSecondChart.data.labels.shift();
        uploadsPerSecondChart.data.datasets[0].data.shift();
    }

    uploadsPerSecondChart.update();

    // Atualizar Uploads por Hora
    uploadsPerHourChart.data.labels.push(currentTime);
    let uploadsInLastHour = 0;
    uploadHistory.forEach(entry => {
        uploadsInLastHour += entry.successes + entry.failures;
    });
    uploadsPerHourChart.data.datasets[0].data.push(uploadsInLastHour);

    // Limitar pontos no gráfico
    const maxDataPointsPerHour = 24; // Últimas 24 horas
    if (uploadsPerHourChart.data.labels.length > maxDataPointsPerHour) {
        uploadsPerHourChart.data.labels.shift();
        uploadsPerHourChart.data.datasets[0].data.shift();
    }

    uploadsPerHourChart.update();
}

// Conectar ao SSE para atualizações em tempo real.
if (!!window.EventSource) {
    let source = new EventSource('/events');

    source.onmessage = function(event) {
        let data = JSON.parse(event.data);
        updateStats(data);
    };

    source.onerror = function(err) {
        console.error("Falha no EventSource:", err);
        source.close();
    };
} else {
    console.log("SSE não é suportado neste navegador.");
}

