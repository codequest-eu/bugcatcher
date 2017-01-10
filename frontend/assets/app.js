(function() {
  Vue.component('error', {
    template: '#error-template',
    props: ['error'],
    data: function() {
      return {
        createdAt: timeSince(this.error.createdAt),
        updatedAt: timeSince(this.error.updatedAt),
        showDetails: false,
      }
    },
    computed: {
      showDetailsCTA: function() {
        return this.showDetails ? 'hide' : 'show';
      }
    },
    methods: {
      toggleDetails: function() {
        this.showDetails = !this.showDetails;
      },
      remove: function() {
        this.$emit('remove', this.error.id);
      }
    }
  });
  Vue.component('event', {
    template: '#event-template',
    props: ['event'],
    data: function() {
      return {
        createdAt: timeSince(this.event.createdAt),
        showMetadata: false,
      }
    },
    computed: {
      toggleMetadataCTA: function() {
        if (this.showMetadata) {
          return 'stack trace';
        }
        return 'metadata';
      },
      prettyMetadata: function() {
        return JSON.stringify(this.event.metaData, null, 2);
      }
    },
    methods: {
      toggleMetadata: function() {
        this.showMetadata = !this.showMetadata;
      }
    }
  });
  Vue.component('stack-frame', {
    template: '#stack-frame-template',
    props: ['frame'],
    data: function() {
      return {
        showContext: false
      }
    },
    methods: {
      toggleContext: function() {
        this.showContext = !this.showContext;
      },
      isHighlighted: function(lineNumber) {
        return (lineNumber == this.frame.lineNumber);
      }
    }
  });
  var app = new Vue({
    el: '#app',
    data: {
      apiKey: 'Loading...',
      errors: []
    },
    mounted: function() {
      axios
        .get('/errors')
        .then(function(response) {
          this.apiKey = response.data.apiKey;
          this.errors = response.data.errors;
        }.bind(this));
    },
    methods: {
      remove: function(id) {
        axios
          .delete('/errors/' + id)
          .then(this.removeFromList(id));
      },
      removeFromList: function(id) {
        return function() {
          this.errors = this.errors.filter(function(error) {
            return error.id != id;
          })
        }.bind(this);
      }
    }
  });
  var timeSince = function(timestamp) {
    var date = new Date(timestamp * 1000);
    var seconds = Math.floor((new Date() - date) / 1000);
    var interval = Math.floor(seconds / 31536000);
    if (interval > 1) {
        return interval + ' years';
    }
    interval = Math.floor(seconds / 2592000);
    if (interval > 1) {
        return interval + ' months';
    }
    interval = Math.floor(seconds / 86400);
    if (interval > 1) {
        return interval + ' days';
    }
    interval = Math.floor(seconds / 3600);
    if (interval > 1) {
        return interval + ' hours';
    }
    interval = Math.floor(seconds / 60);
    if (interval > 1) {
        return interval + ' minutes';
    }
    return Math.floor(seconds) + ' seconds';
  }
})();
