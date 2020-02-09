.POSIX:
SHELL:=/bin/bash
S3_BUCKET=digitalmaneuver.com
OUTPUTDIR=./public
INPUTDIR=./content

upload:

	rm -rf $(OUTPUTDIR)/*

	find $(INPUTDIR) -type f -exec cp {} $(OUTPUTDIR) \;
	
	find $(OUTPUTDIR)  \( -name '*.png' -o -name '*.jpg' -o -name '*.jpeg' \) -print0 | xargs -0 -P8 -n2 mogrify -strip -thumbnail '1000>'
	
	find $(OUTPUTDIR) -type f \( -name "*.css" -or -name "*.html" -or -name "*.js" \) -exec gzip -n "{}" \; -exec mv "{}.gz" "{}" \;
	
	aws s3 sync $(OUTPUTDIR)  s3://$(S3_BUCKET) --metadata-directive REPLACE --acl public-read --content-encoding gzip --content-type "application/javascript; charset=utf-8" --cache-control max-age=3600 --exclude '*' --include '*.js'
	
	aws s3 sync $(OUTPUTDIR)  s3://$(S3_BUCKET) --metadata-directive REPLACE --acl public-read --content-encoding gzip --content-type "text/css; charset=utf-8" --cache-control max-age=3600 --exclude '*' --include '*.css'
	
	aws s3 sync $(OUTPUTDIR)  s3://$(S3_BUCKET) --metadata-directive REPLACE --acl public-read --content-encoding gzip --content-type "text/html; charset=utf-8" --cache-control max-age=3600 --exclude '*' --include '*.html'
	
	aws s3 sync $(OUTPUTDIR) s3://$(S3_BUCKET) --metadata-directive REPLACE --acl public-read --cache-control max-age=3600
	
	aws s3 sync $(OUTPUTDIR)  s3://$(S3_BUCKET) --acl public-read --delete