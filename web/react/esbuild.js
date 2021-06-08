require('esbuild').buildSync({
    entryPoints: ['src/index.jsx'],
    bundle: true,
    minify: true,
    sourcemap: false,
    // target: ['chrome58', 'firefox57', 'safari11', 'edge16'],
    outdir: 'build',
  })