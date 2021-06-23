# Caching image proxy

The idea here is to use `nginx` as a caching proxy to provide image
rescaling functionality. This can also be fronted by a CDN for
performance. I'm going to follow more or less what's described in
[this article][nginx].


## Deployment

The `Dockerfile` here creates an image that reads the server name and
storage bucket name to use from environment variables:

 - `SERVER_NAME`: the name to use in nginx's `server` directive, e.g.
   `img.veganbase.com`.
 - `IMAGE_BUCKET`: the name of the Google Cloud Storage bucket used as
   backing store for images, e.g. `veganbase-images`.


## URL scheme

Images can be downloaded in their original uploaded form using their
bare IDs as the path part of the URL, e.g.

```
GET https://img.veganbase.com/ERTWdfe243rwddsf.jpg
```

Images can be resized by adding query parameters, e.g.

```
GET https://img.veganbase.com/avatar-23.png?width=64
```

The supported parameters are:

 - `width`, `height`: output image size in pixels; if one is omitted,
   the other is either taken from the original image size (for "fill"
   mode) or is scaled to keep the image aspect ratio the same.
 - `quality` (0-100): quality factor for JPEG rescaling.
 - `fill` (true/false): controls the image resizing mode (see below).


## Image resizing modes

There are two different resizing modes, "normal" mode and "fill" mode.
Both modes will only reduce the size of images -- they will *not*
scale images up!

In "normal" mode, the image is resized maintaining its aspect ratio.
The resulting image does not necessarily have exactly the dimensions
specified by the width and height parameters, but has a size where the
largest dimension matches one of the specified size parameters. The
simplest way to use this mode is to specify just one of width or
height and so get back an image of that width or height with the same
aspect ratio as the original image.

In "fill" mode, the output image always has exactly the requested
width and height (defaulting missing values to those taken from the
original image). The input image is scaled to fit the largest
specified output dimension and is then cropped to fit the output image
size.


## Google Cloud Storage backing

 - Create bucket with Bucket Policy only permissions (I think this is
   right, since all of our imagery should be publicly viewable via the
   caching proxy, but is not editable by users).
 - The backing bucket needs to be set up with public Viewer
   permissions: set `allUsers` policy to "Storage Viewer".
 - As well as the public "Storage Viewer" permissions, we'll also need
   a service account with upload permissions for the blob service.
 - If bucket name is "veganbase-image", then images are accessible
   via, e.g http://storage.googleapis.com/veganbase-images/avatar-23.png


[nginx]: http://charlesleifer.com/blog/nginx-a-caching-thumbnailing-reverse-proxying-image-server-/
