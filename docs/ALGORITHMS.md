
# ImageD Algorithms

## Perceptual Hashing

### Average Hash (AHash)
- Resizes image to 8x8
- Converts to grayscale  
- Calculates average pixel value
- Creates 64-bit hash based on above/below average

### Perception Hash (PHash)
- Resizes image to 32x32
- Applies DCT (Discrete Cosine Transform)
- Takes low-frequency 8x8 components
- Creates hash based on coefficient comparisons

### Difference Hash (DHash)  
- Resizes image to 9x8
- Compares adjacent pixels
- Creates hash based on brightness differences

### Wavelet Hash (WHash)
- Uses wavelet transform for scale invariance
- Better for cropped and resized images
- More computationally intensive

## Quality Analysis

### Sharpness Detection
- Uses Laplacian variance method
- Measures high-frequency content
- Higher variance indicates sharper image

### Noise Estimation  
- Analyzes local variance in smooth regions
- Detects high-frequency random variations
- Identifies different noise patterns

### Exposure Assessment
- Calculates luminance histogram
- Analyzes distribution of dark/bright pixels
- Penalizes under/overexposed images

### Color Analysis
- Detects color casts and imbalances
- Analyzes RGB distribution
- Identifies color temperature

## Similarity Comparison

### Hamming Distance
- Used for perceptual hash comparison
- Counts differing bits between hashes
- Fast and efficient for binary data

### Cosine Similarity
- Used for feature vector comparison
- Measures angle between vectors
- Good for high-dimensional data

### Chi-squared Distance
- Used for histogram comparison
- Statistical measure of distribution difference
- Effective for color histograms