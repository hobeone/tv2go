showService = angular.module('tv2go.showsService', ['ngResource']);

showService.factory('Show', function($cacheFactory, $resource) {
    return $resource('http://localhost:9001/api/1/shows/:showid', {}, {
      all: {
        method: "GET",
        isArray:true,
        cache: true,
      },
      update: {
        method: "PUT"
      },
    });
  });

angular.module('tv2go.models.shows',['tv2go.showsService'])
.service('ShowsModel', function($http, $q, Show) {
  var model = this;
  var shows;
  var currentShow;

  function cacheShows(result) {
    shows = result;
    return shows;
  }
  model.getShows = function() {
    return (shows) ? $q.when(shows): Show.all().$promise.then(cacheShows);
  };

  model.setCurrentShow = function(showId) {
    return model.getShowById(showId)
    .then(function(show){
      currentShow = show;
    });
  };

  model.getCurrentShow = function() {
    return currentShow;
  };

  model.getCurrentShowId = function() {
    return currentShow ? currentShow.id : "";
  };

  model.getShowById = function(showId) {
    var deferred = $q.defer();
    function findShow() {
      return _.find(shows, function(s) {
        return s.id === _.parseInt(showId,10);
      });
    }
    if(shows) {
      deferred.resolve(findShow());
    } else {
      model.getShows()
      .then(function(result){
        deferred.resolve(findShow());
      });
    }
    return deferred.promise;
  };
})
;