package gocaption

const sampleDXFP string = `
<?xml version="1.0" encoding="utf-8"?>
<tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml"
    xmlns:tts="http://www.w3.org/ns/ttml#styling">
 <head>
  <styling>
   <style xml:id="p" tts:color="#ffeedd" tts:fontfamily="Arial"
          tts:fontsize="10pt" tts:textAlign="center"/>
  </styling>
  <layout>
  <region tts:displayAlign="after" tts:textAlign="center" xml:id="bottom"></region>
  </layout>
 </head>
 <body>
  <div xml:lang="en-US">
   <p begin="00:00:09.209" end="00:00:12.312" style="p">
    ( clock ticking )
   </p>
   <p begin="00:00:14.848" end="00:00:17.000" style="p">
    MAN:<br/>
    When we think<br/>
    ♪ ...say bow, wow, ♪
   </p>
   <p begin="00:00:17.000" end="00:00:18.752" style="p">
    <span tts:textalign="right">we have this vision of Einstein</span>
   </p>
   <p begin="00:00:18.752" end="00:00:20.887" style="p">
   <br/>
    as an old, wrinkly man<br/>
    with white hair.
   </p>
   <p begin="00:00:20.887" end="00:00:26.760" style="p">
    MAN 2:<br/>
    E equals m c-squared is<br/>
    not about an old Einstein.
   </p>
   <p begin="00:00:26.760" end="00:00:32.200" style="p">
    MAN 2:<br/>
    It's all about an eternal Einstein.
   </p>
   <p begin="00:00:32.200" end="00:00:36.200" style="p">
    &lt;LAUGHING &amp; WHOOPS!&gt;
   </p>
  </div>
 </body>
</tt> `
