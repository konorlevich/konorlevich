const gulp = require('gulp');
const sass = require('gulp-sass');
const sourcemap = require('gulp-sourcemaps');
sass.compiler = require('node-sass');
const minify = require('gulp-minify');
const cleaner = require('gulp-clean');


function clean() {
    return gulp.src('build/*')
        .pipe(cleaner());
}

function styleTranspile() {
    return gulp.src('source/styles/*.scss')
        .pipe(sourcemap.init())
        .pipe(sass({}).on('error', sass.logError))
        .pipe(sourcemap.write())
        .pipe(gulp.dest('build/styles'));
}

function scriptTranspile() {
    return gulp.src('source/scripts/*.js')
        .pipe(gulp.dest('build/scripts'));
}

function styleMinify() {
    let sassOptions = {
        outputStyle: 'compressed'
    };
    return gulp.src('source/styles/*.scss')
        .pipe(sourcemap.init())
        .pipe(sass(sassOptions).on('error', sass.logError))
        .pipe(sourcemap.write())
        .pipe(gulp.dest('build/styles'));
}

function scriptsMinify() {
    let minifyOpts = {
        noSource: true
    };
    return gulp.src('source/scripts/*.js')
        .pipe(sourcemap.init())
        .pipe(minify(minifyOpts))
        .pipe(sourcemap.write())
        .pipe(gulp.dest('build/scripts'));
}

function scriptsWatch() {
    return gulp.watch('source/scripts/*.js', scriptTranspile)
}

function stylesWatch() {
    return gulp.watch('source/styles/*.scss', styleTranspile);
}

exports.watch = gulp.parallel(stylesWatch, scriptsWatch);

exports.build =
    gulp.series(
        clean,
        gulp.parallel(
            styleMinify,
            scriptsMinify,
        ));
exports.default = exports.build;