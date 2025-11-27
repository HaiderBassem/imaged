# ImageD Benchmarks

## Performance Metrics

### Scanning Performance
- **Small collection (1,000 images)**: ~30 seconds
- **Medium collection (10,000 images)**: ~5 minutes  
- **Large collection (100,000 images)**: ~45 minutes

### Memory Usage
- **Base memory**: ~50 MB
- **Per image**: ~2-5 MB (depending on analysis depth)
- **Maximum recommended**: 2,000 images in parallel

### Quality Analysis Speed
- **Basic analysis**: 50-100 ms per image
- **Detailed analysis**: 200-500 ms per image
- **GPU accelerated**: 10-50 ms per image

## Hardware Recommendations

### Minimum Requirements
- CPU: 2+ cores
- RAM: 4 GB
- Storage: 1 GB free space

### Recommended Setup  
- CPU: 4+ cores
- RAM: 8+ GB
- Storage: SSD preferred
- GPU: Optional for acceleration

### Production Deployment
- CPU: 8+ cores
- RAM: 16+ GB
- Storage: Fast SSD
- GPU: NVIDIA with CUDA support

## Optimization Tips

1. **Adjust worker count** based on CPU cores
2. **Use fast scan mode** for initial analysis
3. **Enable GPU** if available
4. **Set memory limits** to prevent swapping
5. **Use appropriate hash algorithms** for your use case