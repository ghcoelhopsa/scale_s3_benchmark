import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import matplotlib.dates as mdates
from datetime import timedelta
from matplotlib.ticker import FuncFormatter
import os

# Configure Seaborn style for better aesthetics
sns.set(style="whitegrid")

# Path to the CSV file
csv_file = 'stats_report.csv'

# Read the CSV file
try:
    data = pd.read_csv(csv_file, parse_dates=['Timestamp'])
except FileNotFoundError:
    print(f"Error: The file '{csv_file}' was not found.")
    exit(1)
except pd.errors.EmptyDataError:
    print(f"Error: The file '{csv_file}' is empty.")
    exit(1)
except pd.errors.ParserError as e:
    print(f"Error parsing the CSV file: {e}")
    exit(1)
except ValueError as e:
    print(f"Error: {e}")
    exit(1)

# Display available columns
print("Available columns in CSV:", data.columns.tolist())

# Check if 'Timestamp' column exists
if 'Timestamp' not in data.columns:
    print("Error: The 'Timestamp' column is not present in the CSV file.")
    exit(1)

# Set 'Timestamp' as the DataFrame index
data.set_index('Timestamp', inplace=True)

# Sort data by index to ensure chronological order
data.sort_index(inplace=True)

# Verify the existence of required columns
required_columns = ['TotalUploads', 'Successes', 'Failures']
for col in required_columns:
    if col not in data.columns:
        print(f"Error: The column '{col}' is not present in the CSV file.")
        exit(1)

# Use 'TotalUploads' directly as 'CumulativeUploads' since it's already cumulative
data['CumulativeUploads'] = data['TotalUploads']

# Calculate per-minute uploads by computing the difference between consecutive entries
data_diff = data['CumulativeUploads'].diff().fillna(0).astype(int)

# Remove any negative values that might occur due to data inconsistencies
data_diff[data_diff < 0] = 0

# Calculate uploads per hour by resampling the data
uploads_per_hour = data_diff.resample('H').sum()

# Calculate additional statistics
total_uploads = data['CumulativeUploads'].iloc[-1]
total_minutes = (data.index[-1] - data.index[0]).total_seconds() / 60
average_per_minute = total_uploads / total_minutes if total_minutes > 0 else 0
average_per_second = average_per_minute / 60  # Estimated
total_hours = total_minutes / 60
average_per_hour = total_uploads / total_hours if total_hours > 0 else 0

# Calculate uploads in the last 10 minutes using .loc with a time mask
end_time = data_diff.index[-1]
start_time_10m = end_time - timedelta(minutes=10)
last_10_minutes = data_diff.loc[start_time_10m:end_time].sum()

# Calculate uploads in the last 24 hours using .loc with a time mask
start_time_24h = end_time - timedelta(hours=24)
last_24_hours = data_diff.loc[start_time_24h:end_time].sum()

# Function to format y-axis labels with K, M, B suffixes
def format_yaxis(x, pos):
    if x >= 1_000_000_000:
        return f'{x/1_000_000_000:.1f}B'
    elif x >= 1_000_000:
        return f'{x/1_000_000:.1f}M'
    elif x >= 1_000:
        return f'{x/1_000:.1f}K'
    else:
        return f'{x}'

# Configure the size of the plot and adjust bottom margin to make space for annotations
plt.figure(figsize=(14, 8))
plt.subplots_adjust(bottom=0.35)  # Adjusted to 0.35 to provide space below

# Plot cumulative uploads
plt.plot(data.index, data['CumulativeUploads'], label='Cumulative Uploads', color='green', linewidth=2)

# Set titles and labels
plt.title('Cumulative S3 Uploads Over Time', fontsize=16)
plt.xlabel('Time', fontsize=14)
plt.ylabel('Cumulative Uploads', fontsize=14)

# Format the x-axis for dates
plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d %H:%M'))
plt.gca().xaxis.set_major_locator(mdates.AutoDateLocator())

# Rotate x-axis labels for better readability
plt.xticks(rotation=45)

# Apply the y-axis formatter
formatter = FuncFormatter(format_yaxis)
plt.gca().yaxis.set_major_formatter(formatter)

# Add legend
plt.legend(loc='upper left')

# Enable grid
plt.grid(True)

# Create the annotation text with the calculated statistics
annotation_text = (
    f"Total Uploads: {total_uploads:,}\n"
    f"Average per Minute: {average_per_minute:,.2f}\n"
    f"Average per Second (Estimated): {average_per_second:,.2f}\n"
    f"Average per Hour: {average_per_hour:,.2f}\n"
    f"Uploads in Last 10 Minutes: {last_10_minutes:,}\n"
    f"Uploads in Last 24 Hours: {last_24_hours:,}"
)

# Add the annotation below the plot using plt.figtext
plt.figtext(0.5, 0.07, annotation_text, ha='center', fontsize=12,
            bbox=dict(facecolor='white', alpha=0.7, edgecolor='gray'))

# Extract date and total uploads for filename
last_date = data.index[-1].strftime('%Y%m%d')
filename = f"plot_{last_date}_{total_uploads}.png"

# Save the plot as a PNG file
plt.savefig(filename)
