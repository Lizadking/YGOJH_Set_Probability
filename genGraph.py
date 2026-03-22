

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

# Read the CSV file with no headers
df = pd.read_csv("data/JUSH-EN035SR/JUSH-EN035SR24.csv", header=None)

# x-axis values are line number 
x = np.arange(1, len(df) + 1)  

# Create the plot
plt.figure(figsize=(10, 6))

# Plot each column as a separate line
for col in df.columns:
    plt.plot(x, df[col], label=f'Column {col + 1}', marker='o', markersize=3)

# Customize the plot
plt.xlabel('Number of Packs')
plt.ylabel('Probability')
plt.title('K9-17 “Ripper” at Super Rare  Binomial Distribution (n = 25m,P(x) = 0.01)')
plt.legend(['Pr[X=x]','Pr[X<=x]','Pr[X>=x]'])
plt.grid(True, alpha=0.3)




plt.tight_layout()
plt.show()
